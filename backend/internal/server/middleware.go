package server

import (
	"log/slog"
	"net/http"
	"time"
)

// SecurityMiddleware adds essential security headers to every response
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware allows frontend applications to access the API safely
func CORSMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		incomingOrigin := r.Header.Get("Origin")
		isAllowed := false
		for _, o := range allowedOrigins {
			slog.Info("CHECKING THIS ORIGIN", "origin", o)
			if o == incomingOrigin {
				isAllowed = true
				break
			}
		}
		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", incomingOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggerMiddleware logs the details and latency of every incoming HTTP request
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", r.RemoteAddr,
		)
	})
}
