package ratelimiter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	_            http.Handler = &httpRateLimiterHandler{}
	_            Extractor    = &httpHeaderExtractor{}
	stateStrings              = map[State]string{
		Allow: "Allow",
		Deny:  "Deny",
	}
)

const (
	rateLimitingTotalRequests = "Rate-Limiting-Total-Requests"
	rateLimitingState         = "Rate-Limiting-State"
	rateLimitingExpiresAt     = "Rate-Limiting-Expires-At"
)

// Extractor represents the way we will extract a key from an HTTP request, this could be
// a value from a header, request path, method used, user authentication information, any information that
// is available at the HTTP request that wouldn't cause side effects if it was collected (this object shouldn't
// read the body of the request).
type Extractor interface {
	Extract(r *http.Request) (string, error)
}

type httpHeaderExtractor struct {
	headers []string
}

// Extract extracts a collection of http headers and joins them to build the key that will be used for
// rate limiting. You should use headers that are guaranteed to be unique for a client.
func (h *httpHeaderExtractor) Extract(r *http.Request) (string, error) {
	values := make([]string, 0, len(h.headers))
	fmt.Println("Values burada: ")
	fmt.Println(values)

	for _, key := range h.headers {
		// if we can't find a value for the headers, give up and return an error.
		if value := strings.TrimSpace(r.Header.Get(key)); value == "" {
			return "", fmt.Errorf("the header %v must have a value set", key)
		} else {
			values = append(values, value)
		}
	}

	return strings.Join(values, "-"), nil
}

/*
// IP extractor implementation
type ipAddressExtractor struct{}

func (e *ipAddressExtractor) Extract(r *http.Request) (string, error) {
	// Try X-Forwarded-For first (if behind proxy)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0]), nil
	}

	// Fall back to RemoteAddr
	ip = r.RemoteAddr
	// RemoteAddr includes port, so strip that off
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	if ip == "" {
		return "", fmt.Errorf("could not extract IP address from request")
	}

	return ip, nil
}
*/
// NewHTTPHeadersExtractor creates a new HTTP header extractor
func NewHTTPHeadersExtractor(headers ...string) Extractor {
	return &httpHeaderExtractor{headers: headers}
	//return &ipAddressExtractor{}
}

// RateLimiterConfig holds the basic config we need to create a middleware http.Handler object that
// performs rate limiting before offloading the request to an actual handler.
type RateLimiterConfig struct {
	Extractor   Extractor
	Strategy    Strategy
	Expiration  time.Duration
	MaxRequests uint64
}

// NewHTTPRateLimiterHandler wraps an existing http.Handler object performing rate limiting before
// sending the request to the wrapped handler. If any errors happen while trying to rate limit a request
// or if the request is denied, the rate limiting handler will send a response to the client and will not
// call the wrapped handler.
func NewHTTPRateLimiterHandler(originalHandler http.Handler, config *RateLimiterConfig) http.Handler {
	return &httpRateLimiterHandler{
		handler: originalHandler,
		config:  config,
	}
}

type httpRateLimiterHandler struct {
	handler http.Handler
	config  *RateLimiterConfig
}

func (h *httpRateLimiterHandler) writeResponse(writer http.ResponseWriter, status int, msg string, args ...interface{}) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(status)
	if _, err := writer.Write([]byte(fmt.Sprintf(msg, args...))); err != nil {
		fmt.Printf("failed to write body to HTTP request: %v", err)
	}
}

// ServeHTTP performs rate limiting with the configuration it was provided and if there were not errors
// and the request was allowed it is sent to the wrapped handler. It also adds rate limiting headers that will be
// sent to the client to make it aware of what state it is in terms of rate limiting.
func (h *httpRateLimiterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	key, err := h.config.Extractor.Extract(request)
	if err != nil {
		h.writeResponse(writer, http.StatusBadRequest, "failed to collect rate limiting key from request: %v", err)
		return
	}
	fmt.Println("KEY BUDUR: ", key)

	result, err := h.config.Strategy.Run(request.Context(), &Request{
		Key:      key,
		Limit:    h.config.MaxRequests,
		Duration: h.config.Expiration,
	})

	if err != nil {
		h.writeResponse(writer, http.StatusInternalServerError, "failed to run rate limiting for request: %v", err)
		return
	}

	// set the rate limiting headers both on allow or deny results so the client knows what is going on
	writer.Header().Set(rateLimitingTotalRequests, strconv.FormatUint(result.TotalRequests, 10))
	writer.Header().Set(rateLimitingState, stateStrings[result.State])
	writer.Header().Set(rateLimitingExpiresAt, result.ExpiresAt.Format(time.RFC3339))

	// when the state is Deny, just return a 429 response to the client and stop the request handling flow
	if result.State == Deny {
		h.writeResponse(writer, http.StatusTooManyRequests, "you have sent too many requests to this service, slow down please")
		return
	}

	// if the request was not denied we assume it was allowed and call the wrapped handler.
	// by leaving this to the end we make sure the wrapped handler is only called once and doesn't have to worry
	// about any rate limiting at all (it doesn't even have to know there was rate limiting happening for this request)
	// as we have already set the headers, so when the handler flushes the response the headers above will be sent.
	h.handler.ServeHTTP(writer, request)
}

/*

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
*/
