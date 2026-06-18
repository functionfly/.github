// Package wasmpoolv1 is the gRPC API for the FunctionFly wasm-pool-service.
//
// The wire format mirrors a hand-rolled protobuf encoding so that
// clients in this repo can compile without running protoc. Production
// deployments regenerate the stubs from wasmpool.proto using the
// standard protoc + protoc-gen-go + protoc-gen-go-grpc toolchain; the
// generated types are wire-compatible with the structs declared here.
//
// Method semantics:
//
//   - Execute: dispatch a single per-tenant request to the replica
//     responsible for the tenant. Returns the JSON output and timing
//     metadata.
//   - Ping: liveness check. Always returns OK with version + timestamp.
//   - Stats: per-replica pool statistics for monitoring/alerting.
package wasmpoolv1

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

// Runtime identifies a supported execution runtime. Mirrors
// github.com/functionfly/wasm.Runtime; the values are stringly typed
// so clients across repos agree on the encoding.
type Runtime string

const (
	RuntimeJavaScript Runtime = "javascript"
	RuntimePython     Runtime = "python"
	RuntimeWASM       Runtime = "wasm"
)

// ExecuteRequest is the per-request payload sent from the orchestrator
// to a pool replica.
type ExecuteRequest struct {
	// TenantID is the tenant the request belongs to. Required.
	TenantID string
	// Runtime is one of "javascript", "python", or "wasm".
	Runtime Runtime
	// WasmPath is the path to a precompiled .wasm module on the pool
	// node's filesystem, used for module-based deployments.
	WasmPath string
	// Code is the source bytes (JavaScript/Python) or raw .wasm bytes
	// for inline deployments.
	Code []byte
	// Input is the per-request JSON input, passed verbatim to the
	// function as its `data` argument.
	Input []byte
	// TimeoutMs mirrors Timeout in milliseconds.
	TimeoutMs uint32
	// Timeout is the per-request deadline. Zero means use the server
	// default (currently 55s).
	Timeout time.Duration
	// MemoryMB is the per-instance memory budget, in MiB.
	MemoryMB uint32
	// FunctionID identifies the function being executed, used for
	// metrics and tracing.
	FunctionID string
	// Version is the function version label.
	Version string
}

// ExecuteResponse is the result of an Execute call.
type ExecuteResponse struct {
	// Output is the JSON output produced by the function.
	Output []byte
	// Error is the user-visible error string, populated on failure.
	Error string
	// LatencyMs is the server-side execution latency in milliseconds.
	LatencyMs uint32
	// MemoryBytes is the high-water memory usage observed during
	// execution.
	MemoryBytes int64
	// ColdStarted is true if the replica had to instantiate a new
	// pool instance for this request.
	ColdStarted bool
}

// PoolStatsRequest asks the replica for its current stats.
type PoolStatsRequest struct{}

// PoolStatsResponse is the per-replica stats payload.
type PoolStatsResponse struct {
	AvailableInstances int32
	InUseInstances     int32
	TenantBuckets      int32
	TotalExecutions    uint64
	UptimeSeconds      uint64
}

// PingRequest is used for health checks.
type PingRequest struct{}

// PingResponse is the matching payload.
type PingResponse struct {
	OK        bool
	Version   string
	Timestamp time.Time
}

// WasmPoolServer is the interface implemented by pool replicas.
type WasmPoolServer interface {
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	Ping(ctx context.Context, req *PingRequest) (*PingResponse, error)
	Stats(ctx context.Context, req *PoolStatsRequest) (*PoolStatsResponse, error)
}

// WasmPoolClient is the interface used by the orchestrator to dispatch
// requests. It is satisfied by both the production gRPC client and any
// in-process implementation (useful for tests).
type WasmPoolClient interface {
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	Ping(ctx context.Context, req *PingRequest) (*PingResponse, error)
	Stats(ctx context.Context, req *PoolStatsRequest) (*PoolStatsResponse, error)
	Close() error
}

// ServiceDescription is the gRPC service descriptor. It is used by the
// production server to register handlers.
var ServiceDescription = &grpc.ServiceDesc{
	ServiceName: "functionfly.wasmpool.v1.WasmPool",
	HandlerType: (*WasmPoolServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Execute", Handler: executeHandler},
		{MethodName: "Ping", Handler: pingHandler},
		{MethodName: "Stats", Handler: statsHandler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "wasmpool/v1/wasmpool.proto",
}

func executeHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WasmPoolServer).Execute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/functionfly.wasmpool.v1.WasmPool/Execute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WasmPoolServer).Execute(ctx, req.(*ExecuteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func pingHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WasmPoolServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/functionfly.wasmpool.v1.WasmPool/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WasmPoolServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func statsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PoolStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WasmPoolServer).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/functionfly.wasmpool.v1.WasmPool/Stats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WasmPoolServer).Stats(ctx, req.(*PoolStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RegisterServer registers the WasmPool service with a gRPC server.
func RegisterServer(s *grpc.Server, srv WasmPoolServer) {
	s.RegisterService(ServiceDescription, srv)
}

// ClientOption configures NewClient.
type ClientOption func(*clientOptions)

type clientOptions struct {
	timeout time.Duration
}

func defaultClientOptions() clientOptions {
	return clientOptions{timeout: 5 * time.Second}
}

// WithTimeout overrides the default per-call timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) { o.timeout = d }
}

// NewWasmPoolClient is an alias for NewClient provided for compatibility
// with callers that expect the proto-generated name.
func NewWasmPoolClient(conn *grpc.ClientConn, opts ...ClientOption) WasmPoolClient {
	return NewClient(conn, opts...)
}

// NewClient returns a gRPC-backed WasmPoolClient.
func NewClient(conn *grpc.ClientConn, opts ...ClientOption) WasmPoolClient {
	o := defaultClientOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &grpcClient{conn: conn, opts: o}
}

type grpcClient struct {
	conn *grpc.ClientConn
	opts clientOptions
}

func (c *grpcClient) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil execute request")
	}
	if c.conn == nil {
		return nil, fmt.Errorf("nil grpc connection")
	}
	if _, ok := ctx.Deadline(); !ok && c.opts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.opts.timeout)
		defer cancel()
	}
	out := new(ExecuteResponse)
	if err := c.conn.Invoke(ctx, "/functionfly.wasmpool.v1.WasmPool/Execute", req, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcClient) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("nil grpc connection")
	}
	out := new(PingResponse)
	if err := c.conn.Invoke(ctx, "/functionfly.wasmpool.v1.WasmPool/Ping", req, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcClient) Stats(ctx context.Context, req *PoolStatsRequest) (*PoolStatsResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("nil grpc connection")
	}
	out := new(PoolStatsResponse)
	if err := c.conn.Invoke(ctx, "/functionfly.wasmpool.v1.WasmPool/Stats", req, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
