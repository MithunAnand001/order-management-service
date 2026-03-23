package middleware

import (
	"net/http"
	"order-management-service/internal/utils"
	"sync"
	"time"
)

// SecurityHeaders adds standard security headers to prevent XSS, Clickjacking, etc.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements a basic in-memory rate limiter
type RateLimiter struct {
	sync.Mutex
	ips map[string][]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{ips: make(map[string][]time.Time)}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr // Simplified; use X-Forwarded-For in prod
		rl.Lock()
		now := utils.Now()

		// Keep only requests from the last minute
		times := []time.Time{}
		for _, t := range rl.ips[ip] {
			if now.Sub(t) < time.Minute {
				times = append(times, t)
			}
		}

		if len(times) >= 60 { // Max 60 requests per minute
			rl.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		times = append(times, now)
		rl.ips[ip] = times
		rl.Unlock()

		next.ServeHTTP(w, r)
	})
}
