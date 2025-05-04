package ratelimiter

import (
	"fmt"
	"net/http"
	"time"
)

// KeyFunc defines a function that extracts a key from an HTTP request
type KeyFunc func(*http.Request) string

// IPKeyFunc returns a KeyFunc that uses the client's IP address
func IPKeyFunc(useXForwardedFor bool) KeyFunc {
	return func(r *http.Request) string {
		var ip string

		if useXForwardedFor {
			ip = r.Header.Get("X-Forwarded-For")
		}

		if ip == "" {
			ip = r.RemoteAddr
		}

		return "ip:" + ip
	}
}

// Middleware returns an HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(keyFn KeyFunc, rate ...Rate) func(http.HandlerFunc) http.HandlerFunc {
	// Use default rate if none provided
	r := rl.defaultRate
	if len(rate) > 0 {
		r = rate[0]
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			// Generate key from request
			key := keyFn(req)

			// Apply rate limiting
			allowed, remaining, err := rl.Allow(req.Context(), key, r)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", r.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(r.Window).Unix()))

			if err != nil {
				// In case of Redis errors, you might want to allow the request
				// or implement a fallback strategy
				http.Error(w, "Rate limiting error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", r.Window.Seconds()))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		}
	}
}
