package advanced_security

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           string // "closed", "open", "half-open"
	failureCount    int
	successCount    int
	nextAttempt     time.Time
	threshold       float64
	timeout         time.Duration
	halfOpenMaxRequests int
}

// RequestQueue manages request queuing for traffic management
type RequestQueue struct {
	mu      sync.RWMutex
	queue   chan *QueuedRequest
	workers int
	timeout time.Duration
}

// CircuitBreaker implementation
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case "closed":
		return true
	case "open":
		if time.Now().After(cb.nextAttempt) {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = "half-open"
			cb.mu.Unlock()
			cb.mu.RLock()
		}
		fallthrough
	case "half-open":
		return cb.successCount < cb.halfOpenMaxRequests
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	if cb.state == "half-open" && cb.successCount >= cb.halfOpenMaxRequests {
		cb.reset()
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	failureRate := float64(cb.failureCount) / float64(cb.failureCount + cb.successCount)

	if failureRate > cb.threshold {
		cb.trip()
	}
}

func (cb *CircuitBreaker) reset() {
	cb.state = "closed"
	cb.failureCount = 0
	cb.successCount = 0
}

func (cb *CircuitBreaker) trip() {
	cb.state = "open"
	cb.nextAttempt = time.Now().Add(cb.timeout)
}

// RequestQueue implementation
func (rq *RequestQueue) QueueRequest(w http.ResponseWriter, r *http.Request, handler http.HandlerFunc) {
	// Skip queuing for WebSocket upgrade requests as they require the original response writer
	if r.Header.Get("Upgrade") == "websocket" {
		handler(w, r)
		return
	}

	select {
	case rq.queue <- &QueuedRequest{
		request:  r,
		response: w,
		handler:  handler,
		done:     make(chan bool, 1),
		started:  time.Now(),
	}:
		// Request queued successfully
	default:
		// Queue is full
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}
}

func (rq *RequestQueue) processQueue() {
	for i := 0; i < rq.workers; i++ {
		go func() {
			for req := range rq.queue {
				if time.Since(req.started) > rq.timeout {
					http.Error(req.response, "Request timeout", http.StatusRequestTimeout)
					continue
				}

				// Process the queued request with context indicating it was queued
				queuedCtx := context.WithValue(req.request.Context(), "queued", true)
				queuedRequest := req.request.WithContext(queuedCtx)

				// Execute the handler function with the queued request
				req.handler(req.response, queuedRequest)

				// Signal completion if needed
				select {
				case req.done <- true:
				default:
				}
			}
		}()
	}
}