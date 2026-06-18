package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"hash/fnv"
	"os"
	"sync"
	"time"

	wasmpoolv1 "github.com/functionfly/functionfly/internal/wasmpoolservice/api/wasmpool/v1"
	"github.com/buraksezer/consistent"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ExternalPoolClient is a gRPC client that talks to the wasm-pool-service
// over a consistent-hash ring on tenantID.
type ExternalPoolClient struct {
	addr    string // headless service DNS name, e.g. "wasm-pool-service:8084"
	auth    string // HMAC token, optional
	tlsCfg  *tls.Config

	mu       sync.RWMutex
	ring     *consistent.Consistent
	conns    map[string]*grpc.ClientConn // addr → conn
	endpoints map[string]struct{}         // set of addrs in the ring
}

// ExternalConfig configures the external client.
type ExternalConfig struct {
	// Addr is the headless service DNS name. The client will resolve it
	// periodically and treat each A record as a pool replica.
	Addr string

	// AuthToken is the HMAC shared secret for dev. Empty disables.
	AuthToken string

	// TLS enables mTLS. When true, CertFile/KeyFile/CAFile are required.
	TLS       bool
	CertFile  string
	KeyFile   string
	CAFile    string
}

// NewExternalPoolClient builds a client and does the first DNS resolve.
func NewExternalPoolClient(cfg ExternalConfig) (*ExternalPoolClient, error) {
	c := &ExternalPoolClient{
		addr:      cfg.Addr,
		auth:      cfg.AuthToken,
		conns:     make(map[string]*grpc.ClientConn),
		endpoints: make(map[string]struct{}),
	}
	if cfg.TLS {
		tlsCfg, err := loadClientTLS(cfg)
		if err != nil {
			return nil, err
		}
		c.tlsCfg = tlsCfg
	}
	if err := c.refreshRing(context.Background()); err != nil {
		return nil, fmt.Errorf("initial ring resolve: %w", err)
	}
	return c, nil
}

// Name returns the transport identifier.
func (c *ExternalPoolClient) Name() string { return "external" }

// Close closes all gRPC connections.
func (c *ExternalPoolClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, conn := range c.conns {
		_ = conn.Close()
	}
	c.conns = nil
	return nil
}

// RefreshEndpoints re-resolves the headless service and rebuilds the ring.
// Safe to call periodically (the plan recommends every 30 s).
func (c *ExternalPoolClient) RefreshEndpoints(ctx context.Context) error {
	return c.refreshRing(ctx)
}

func (c *ExternalPoolClient) refreshRing(ctx context.Context) error {
	addrs, err := resolveHeadless(ctx, c.addr)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	newSet := make(map[string]struct{}, len(addrs))
	for _, a := range addrs {
		newSet[a] = struct{}{}
	}

	if c.ring == nil {
		// First-time build.
		cfg := consistent.Config{
			PartitionCount:    271,
			ReplicationFactor: 20,
			Load:              1.25,
			Hasher:            fnvHasher{},
		}
		members := make([]consistent.Member, 0, len(addrs))
		for _, a := range addrs {
			members = append(members, StringMember(a))
		}
		c.ring = consistent.New(members, cfg)
	} else {
		// Incremental update.
		for a := range c.endpoints {
			if _, ok := newSet[a]; !ok {
				c.ring.Remove(a)
			}
		}
		for _, a := range addrs {
			if _, ok := c.endpoints[a]; !ok {
				c.ring.Add(StringMember(a))
			}
		}
	}

	// Close connections to removed endpoints.
	for addr, conn := range c.conns {
		if _, ok := newSet[addr]; !ok {
			_ = conn.Close()
			delete(c.conns, addr)
		}
	}
	c.endpoints = newSet
	return nil
}

// Endpoints returns the current set of resolved endpoints (for tests).
func (c *ExternalPoolClient) Endpoints() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, 0, len(c.endpoints))
	for a := range c.endpoints {
		out = append(out, a)
	}
	return out
}

// OwnerOf returns the endpoint currently responsible for tenantID.
func (c *ExternalPoolClient) OwnerOf(tenantID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.ring == nil {
		return ""
	}
	m := c.ring.LocateKey([]byte(tenantID))
	if m == nil {
		return ""
	}
	return m.String()
}

// Execute dispatches a request to the replica responsible for tenantID.
func (c *ExternalPoolClient) Execute(ctx context.Context, req *Request) (*Response, error) {
	owner := c.OwnerOf(req.TenantID)
	if owner == "" {
		return nil, status.Error(codes.Unavailable, "no pool replicas available")
	}
	conn, err := c.connFor(owner)
	if err != nil {
		return nil, err
	}
	if c.auth != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-pool-token", c.auth)
	}
	cli := wasmpoolv1.NewWasmPoolClient(conn)
	gresp, err := cli.Execute(ctx, &wasmpoolv1.ExecuteRequest{
		TenantID:   req.TenantID,
		Runtime:    wasmpoolv1.Runtime(req.Runtime),
		WasmPath:   req.WasmPath,
		Input:      req.Input,
		TimeoutMs:  uint32(req.Timeout / time.Millisecond),
		MemoryMB:   req.MemoryMB,
		FunctionID: req.FunctionID,
		Version:    req.Version,
	})
	if err != nil {
		return nil, err
	}
	resp := &Response{
		Output:      gresp.Output,
		Latency:     time.Duration(gresp.LatencyMs) * time.Millisecond,
		MemoryBytes: uint64(gresp.MemoryBytes),
		ColdStarted: gresp.ColdStarted,
	}
	if gresp.Error != "" {
		resp.Error = gresp.Error
	}
	return resp, nil
}

func (c *ExternalPoolClient) connFor(addr string) (*grpc.ClientConn, error) {
	c.mu.RLock()
	if conn, ok := c.conns[addr]; ok {
		c.mu.RUnlock()
		return conn, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if conn, ok := c.conns[addr]; ok {
		return conn, nil
	}
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(8<<20),
			grpc.MaxCallSendMsgSize(8<<20),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
		}),
		grpc.WithBlock(),
		grpc.WithReturnConnectionError(),
	}
	if c.tlsCfg != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsCfg)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	c.conns[addr] = conn
	return conn, nil
}

func loadClientTLS(cfg ExternalConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %w", err)
	}
	pool := x509.NewCertPool()
	if cfg.CAFile != "" {
		caBytes, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA: %w", err)
		}
		if !pool.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("failed to append CA certs")
		}
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// StringMember is a minimal consistent.Member that just carries the addr.
type StringMember string

func (m StringMember) String() string { return string(m) }

// fnvHasher is a simple FNV-1a hasher for the consistent hash ring. The
// buraksezer/consistent library requires Hasher to be set in Config.
type fnvHasher struct{}

func (fnvHasher) Sum64(b []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(b)
	return h.Sum64()
}

// resolveHeadless does a DNS A-record lookup. Uses the stdlib net.Resolver
// to keep the dependency surface small.
func resolveHeadless(ctx context.Context, addr string) ([]string, error) {
	host, port, err := splitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := netLookupHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup %s: %w", host, err)
	}
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, joinHostPort(ip, port))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no endpoints for %s", host)
	}
	return out, nil
}

// LogEndpoints logs the current endpoint set, for ops visibility.
func (c *ExternalPoolClient) LogEndpoints() {
	addrs := c.Endpoints()
	logrus.WithField("endpoints", addrs).Info("wasm pool: external endpoints")
}
