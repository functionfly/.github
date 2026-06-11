package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

type RequestIDMiddleware struct {
	headerName string
}

func NewRequestIDMiddleware(headerName string) *RequestIDMiddleware {
	if headerName == "" {
		headerName = "X-Request-ID"
	}
	return &RequestIDMiddleware{headerName: headerName}
}

func (m *RequestIDMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(m.headerName)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set(m.headerName, requestID)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func GetRequestIDFromRequest(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok {
		return id
	}
	return r.Header.Get("X-Request-ID")
}

func RequestIDHandler(next http.HandlerFunc) http.HandlerFunc {
	return NewRequestIDMiddleware("X-Request-ID").Handler(next)
}

type RequestIDResponseWriter struct {
	http.ResponseWriter
	requestID string
}

func NewRequestIDResponseWriter(w http.ResponseWriter, requestID string) *RequestIDResponseWriter {
	return &RequestIDResponseWriter{
		ResponseWriter: w,
		requestID:       requestID,
	}
}

func (w *RequestIDResponseWriter) WriteHeader(statusCode int) {
	w.Header().Set("X-Request-ID", w.requestID)
	w.ResponseWriter.WriteHeader(statusCode)
}

type RequestIDMiddlewareWithResponse struct {
	headerName string
}

func NewRequestIDMiddlewareWithResponse() *RequestIDMiddlewareWithResponse {
	return &RequestIDMiddlewareWithResponse{headerName: "X-Request-ID"}
}

func (m *RequestIDMiddlewareWithResponse) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(m.headerName)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set(m.headerName, requestID)

		rw := &requestIDWriter{ResponseWriter: w, requestID: requestID}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), RequestIDKey, requestID)))
	}
}

type requestIDWriter struct {
	http.ResponseWriter
	requestID string
}

func (w *requestIDWriter) WriteHeader(statusCode int) {
	w.Header().Set("X-Request-ID", w.requestID)
	w.ResponseWriter.WriteHeader(statusCode)
}

func AddRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetOrCreateRequestID(r *http.Request) string {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return requestID
}