// Command wasm-pool-load is a synthetic load test for the wasm-pool-service.
// It exercises the gRPC Execute endpoint with N tenants at a target RPS
// and reports p50 / p95 / p99 latency + cold-start count.
//
// Usage:
//
//	go run ./tests/load/wasm-pool \
//	  -addr wasm-pool-service:8084 \
//	  -rps 100 \
//	  -tenants 50 \
//	  -duration 10m
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	wasmpoolv1 "github.com/functionfly/functionfly/internal/wasmpoolservice/api/wasmpool/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:8084", "wasm-pool-service gRPC address")
	rps := flag.Int("rps", 100, "target requests per second")
	tenants := flag.Int("tenants", 50, "number of simulated tenants")
	duration := flag.Duration("duration", 10*time.Minute, "test duration")
	payload := flag.Int("payload", 256, "input payload size in bytes")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	cli := wasmpoolv1.NewWasmPoolClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	const reportEvery = 30 * time.Second
	latencies := &latencyRing{cap: int(*rps) * 5}
	var total, errors, coldStarts atomic.Int64

	var wg sync.WaitGroup
	workers := *rps / 10
	if workers < 1 {
		workers = 1
	}
	tokens := make(chan struct{}, workers)
	ticker := time.NewTicker(time.Second / time.Duration(*rps/workers+1))
	defer ticker.Stop()
	reportTicker := time.NewTicker(reportEvery)
	defer reportTicker.Stop()

	log.Printf("starting: addr=%s rps=%d tenants=%d duration=%s", *addr, *rps, *tenants, *duration)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			printReport(os.Stdout, total.Load(), errors.Load(), coldStarts.Load(), latencies.snapshot())
			return
		case <-reportTicker.C:
			printReport(os.Stdout, total.Load(), errors.Load(), coldStarts.Load(), latencies.snapshot())
		case <-ticker.C:
			tokens <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-tokens }()
				start := time.Now()
				tenantID := fmt.Sprintf("tenant-%d", rand.Intn(*tenants))
				input := make([]byte, *payload)
				resp, err := cli.Execute(ctx, &wasmpoolv1.ExecuteRequest{
					TenantID: tenantID,
					Runtime:  "python",
					Input:    input,
					TimeoutMs: 5000,
				})
				latencies.add(time.Since(start))
				total.Add(1)
				if err != nil || (resp != nil && resp.Error != "") {
					errors.Add(1)
					return
				}
				if resp.ColdStarted {
					coldStarts.Add(1)
				}
			}()
		}
	}
}

type latencyRing struct {
	mu   sync.Mutex
	data []time.Duration
	cap  int
}

func (l *latencyRing) add(d time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.data) >= l.cap {
		// drop oldest
		l.data = l.data[1:]
	}
	l.data = append(l.data, d)
}

func (l *latencyRing) snapshot() []time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]time.Duration, len(l.data))
	copy(out, l.data)
	return out
}

func printReport(w *os.File, total, errors, cold int64, latencies []time.Duration) {
	if total == 0 {
		fmt.Fprintf(w, "[%s] no requests yet\n", time.Now().Format(time.RFC3339))
		return
	}
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	p := func(q float64) time.Duration {
		if len(sorted) == 0 {
			return 0
		}
		idx := int(float64(len(sorted)-1) * q)
		return sorted[idx]
	}
	fmt.Fprintf(w, "[%s] total=%d errors=%d (%.2f%%) cold=%d p50=%s p95=%s p99=%s\n",
		time.Now().Format(time.RFC3339),
		total, errors, float64(errors)/float64(total)*100, cold,
		p(0.50), p(0.95), p(0.99))
}
