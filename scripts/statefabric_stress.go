package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL  = "http://localhost:8080"
	duration = 30 * time.Second
	workers  = 20
)

func getEnvOrFail(key, description string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "FATAL: %s (set via %s environment variable)\n", description, key)
		os.Exit(1)
	}
	return val
}

type Stats struct {
	total       int64
	success     int64
	clientErr   int64
	serverErr   int64
	totalLatNs  int64
	maxLatNs    int64
	minLatNs    int64
	statusCodes sync.Map
}

func (s *Stats) record(status int, dur time.Duration) {
	atomic.AddInt64(&s.total, 1)
	ns := dur.Nanoseconds()
	atomic.AddInt64(&s.totalLatNs, ns)
	for {
		old := atomic.LoadInt64(&s.maxLatNs)
		if ns <= old || atomic.CompareAndSwapInt64(&s.maxLatNs, old, ns) {
			break
		}
	}
	if s.minLatNs == 0 || ns < s.minLatNs {
		atomic.StoreInt64(&s.minLatNs, ns)
	}
	v, _ := s.statusCodes.LoadOrStore(status, new(int64))
	atomic.AddInt64(v.(*int64), 1)
	switch {
	case status >= 200 && status < 300:
		atomic.AddInt64(&s.success, 1)
	case status >= 400 && status < 500:
		atomic.AddInt64(&s.clientErr, 1)
	case status >= 500:
		atomic.AddInt64(&s.serverErr, 1)
	}
}

func login() (string, error) {
	email := getEnvOrFail("STRESS_TEST_EMAIL", "email is required for stress test")
	password := getEnvOrFail("STRESS_TEST_PASSWORD", "password is required for stress test")
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := http.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	tok, _ := v["token"].(string)
	return tok, nil
}

func doRequest(client *http.Client, method, path, token string, body []byte) (int, time.Duration) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, _ := http.NewRequest(method, baseURL+path, rdr)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	start := time.Now()
	resp, err := client.Do(req)
	dur := time.Since(start)
	if err != nil {
		return 0, dur
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, dur
}

func percentile(latencies []int64, p float64) int64 {
	if len(latencies) == 0 {
		return 0
	}
	idx := int(float64(len(latencies)) * p)
	if idx >= len(latencies) {
		idx = len(latencies) - 1
	}
	return latencies[idx]
}

func main() {
	fmt.Println("=== State Fabric Stress Test ===")
	fmt.Printf("Target: %s\n", baseURL)
	fmt.Printf("Workers: %d, Duration: %s\n", workers, duration)

	token, err := login()
	if err != nil || token == "" {
		fmt.Fprintf(os.Stderr, "FATAL: login failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Auth: OK")

	client := &http.Client{Timeout: 10 * time.Second}

	fabricBody, _ := json.Marshal(map[string]string{"name": fmt.Sprintf("stress-%d", time.Now().UnixNano()), "description": "stress", "fabric_type": "custom"})

	phases := []struct {
		name   string
		weight int
		fn     func() (int, time.Duration)
	}{
		{"health", 3, func() (int, time.Duration) { return doRequest(client, "GET", "/health", "", nil) }},
		{"list_fabrics", 3, func() (int, time.Duration) { return doRequest(client, "GET", "/v1/state-fabrics", token, nil) }},
		{"create_fabric", 1, func() (int, time.Duration) { return doRequest(client, "POST", "/v1/state-fabrics", token, fabricBody) }},
		{"feature_flags", 2, func() (int, time.Duration) { return doRequest(client, "GET", "/v1/state-fabrics/feature-flags", token, nil) }},
		{"admin_stats", 1, func() (int, time.Duration) { return doRequest(client, "GET", "/v1/admin/state-fabrics/stats", token, nil) }},
	}

	stats := &Stats{minLatNs: 0}
	var latencies []int64
	var latMu sync.Mutex

	totalWeight := 0
	for _, p := range phases {
		totalWeight += p.weight
	}

	deadline := time.Now().Add(duration)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for time.Now().Before(deadline) {
				// pick phase by weight
				r := time.Now().UnixNano() % int64(totalWeight)
				var phase func() (int, time.Duration)
				acc := int64(0)
				for _, p := range phases {
					acc += int64(p.weight)
					if r < acc {
						phase = p.fn
						break
					}
				}
				if phase == nil {
					continue
				}
				status, dur := phase()
				stats.record(status, dur)
				latMu.Lock()
				latencies = append(latencies, dur.Nanoseconds())
				latMu.Unlock()
			}
		}(w)
	}

	// progress ticker
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				total := atomic.LoadInt64(&stats.total)
				fmt.Printf("  [%s] requests=%d success=%d 4xx=%d 5xx=%d\n",
					time.Now().Format("15:04:05"), total,
					atomic.LoadInt64(&stats.success),
					atomic.LoadInt64(&stats.clientErr),
					atomic.LoadInt64(&stats.serverErr))
			}
		}
	}()

	wg.Wait()
	close(done)

	total := atomic.LoadInt64(&stats.total)
	elapsed := duration.Seconds()
	rps := float64(total) / elapsed
	avgLat := time.Duration(0)
	if total > 0 {
		avgLat = time.Duration(atomic.LoadInt64(&stats.totalLatNs) / total)
	}
	maxLat := time.Duration(atomic.LoadInt64(&stats.maxLatNs))
	minLat := time.Duration(atomic.LoadInt64(&stats.minLatNs))

	latMu.Lock()
	// simple sort for percentiles
	for i := 1; i < len(latencies); i++ {
		for j := i; j > 0 && latencies[j] < latencies[j-1]; j-- {
			latencies[j], latencies[j-1] = latencies[j-1], latencies[j]
		}
	}
	p50 := time.Duration(percentile(latencies, 0.50))
	p95 := time.Duration(percentile(latencies, 0.95))
	p99 := time.Duration(percentile(latencies, 0.99))
	latMu.Unlock()

	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("Total requests:   %d\n", total)
	fmt.Printf("Duration:         %s\n", duration)
	fmt.Printf("Throughput:       %.1f req/s\n", rps)
	fmt.Printf("Success (2xx):    %d (%.1f%%)\n", atomic.LoadInt64(&stats.success), float64(atomic.LoadInt64(&stats.success))/float64(total)*100)
	fmt.Printf("Client err (4xx): %d (%.1f%%)\n", atomic.LoadInt64(&stats.clientErr), float64(atomic.LoadInt64(&stats.clientErr))/float64(total)*100)
	fmt.Printf("Server err (5xx): %d (%.1f%%)\n", atomic.LoadInt64(&stats.serverErr), float64(atomic.LoadInt64(&stats.serverErr))/float64(total)*100)
	fmt.Println()
	fmt.Println("Latency:")
	fmt.Printf("  min:  %s\n", minLat)
	fmt.Printf("  avg:  %s\n", avgLat)
	fmt.Printf("  p50:  %s\n", p50)
	fmt.Printf("  p95:  %s\n", p95)
	fmt.Printf("  p99:  %s\n", p99)
	fmt.Printf("  max:  %s\n", maxLat)
	fmt.Println()
	fmt.Println("Status code distribution:")
	stats.statusCodes.Range(func(k, v interface{}) bool {
		fmt.Printf("  %d: %d\n", k.(int), atomic.LoadInt64(v.(*int64)))
		return true
	})

	// Pass/fail criteria
	errorRate := float64(atomic.LoadInt64(&stats.serverErr)) / float64(total) * 100
	if errorRate > 5.0 {
		fmt.Fprintf(os.Stderr, "\nFAIL: server error rate %.1f%% exceeds 5%% threshold\n", errorRate)
		os.Exit(1)
	}
	fmt.Println("\nPASS: server error rate within 5% threshold")
}
