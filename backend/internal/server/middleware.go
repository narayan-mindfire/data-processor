package server

import (
	"golang.org/x/time/rate"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

// RateLimiterMiddleware blocks IP addresses that exceed the allowed request rate
func RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		limiter := getVisitor(ip)
		if !limiter.Allow() {
			// If they exceed 5 requests/secs
			http.Error(w, "429 Too Many Requests - Please slow down!", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getVisitor retrieves or creates a rate limiter for a specific IP
func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	limiter, exists := visitors[ip]
	if !exists {
		// Allow 5 requests per second, with a maximum burst of 10
		limiter = rate.NewLimiter(5, 10)
		visitors[ip] = limiter
	}
	return limiter
}

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
