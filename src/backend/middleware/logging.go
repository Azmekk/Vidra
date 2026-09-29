package middleware

import (
	"bufio"
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// LogFailures writes one log line for every API request that ends with a 4xx
// or 5xx status, including the error sent to the client and any annotations.
func LogFailures(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/swagger/") {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.status < http.StatusBadRequest {
			return
		}
		level := slog.LevelWarn
		if rec.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		attrs := append([]any{
			"method", r.Method, "path", r.URL.Path, "status", rec.status,
			"ip", chimw.GetClientIP(r.Context()), "took", time.Since(start).Round(time.Millisecond),
		}, rec.attrs...)
		slog.Log(r.Context(), level, "request failed", attrs...)
	})
}

// ClientIPFrom returns the client address resolved by ClientIP.
func ClientIPFrom(ctx context.Context) string { return chimw.GetClientIP(ctx) }

type recorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	attrs       []any
}

func (r *recorder) Annotate(args ...any) { r.attrs = append(r.attrs, args...) }

func (r *recorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status, r.wroteHeader = code, true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

func (r *recorder) Flush() { _ = http.NewResponseController(r.ResponseWriter).Flush() }

func (r *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(r.ResponseWriter).Hijack()
}

func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
