package a2a

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// StreamSSE writes Server-Sent Events to the response writer.
// It blocks until the context is cancelled or the channel is closed.
func StreamSSE(w http.ResponseWriter, flusher http.Flusher, ch <-chan SSEEvent, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(data))
			flusher.Flush()
		case <-time.After(30 * time.Second):
			// Send a keep-alive comment every 30 seconds.
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
