package clientx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	mu sync.RWMutex
	// Custom HTTP client
	defaultClient = &http.Client{
		Transport: &http.Transport{
			// Automatically read proxy from environment variables, e.g., HTTP_PROXY / HTTPS_PROXY
			Proxy: http.ProxyFromEnvironment,

			// Maximum number of idle connections globally, suitable for high concurrency scenarios
			MaxIdleConns: 100,

			// Idle connection timeout, close if not used for this duration to release resources
			IdleConnTimeout: 90 * time.Second,

			// TLS handshake timeout, report error if handshake not completed within this time
			TLSHandshakeTimeout: 10 * time.Second,

			// HTTP/1.1 Expect: 100-continue timeout
			ExpectContinueTimeout: 1 * time.Second,

			// Maximum number of idle connections per host, dynamically set based on CPU cores
			MaxIdleConnsPerHost: runtime.GOMAXPROCS(0) + 1,

			// TLS configuration
			TLSClientConfig: &tls.Config{
				// Skip certificate verification
				InsecureSkipVerify: true,
				// For security, you can use a custom CA:
				// RootCAs: x509.NewCertPool()
			},

			// Whether to disable Keep-Alive, false means enable TCP connection reuse for better performance
			DisableKeepAlives: false,
		},

		// Request timeout (including connection, sending request, reading response)
		Timeout: 10 * time.Second,
	}
	// bufferPool for reusing large request bodies
	bufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

// Middleware defines middleware type that can execute logic before/after requests
type Middleware func(next func(*http.Request) (*http.Response, error)) func(*http.Request) (*http.Response, error)

// SetClient replaces the global HTTP client
func SetClient(client *http.Client) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient = client
}

// GetClient gets the global HTTP client
func GetClient() *http.Client {
	mu.RLock()
	defer mu.RUnlock()
	return defaultClient
}

// Option request configuration structure
type Option struct {
	Retries     int               // Maximum number of retries
	Backoff     BackoffFunc       // Retry backoff strategy
	Headers     map[string]string // Custom request headers
	ForceRetry  bool              // Whether to force retry for all methods
	Middlewares []Middleware      // Middleware chain
}

// BackoffFunc defines retry backoff function
type BackoffFunc func(attempt int) time.Duration

// Default exponential backoff, max 30s
func defaultBackoff(attempt int) time.Duration {
	d := time.Duration(1<<attempt) * 500 * time.Millisecond
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}

// OptionFunc functional configuration type
type OptionFunc func(*Option)

// WithRetries sets maximum number of retries
func WithRetries(n int) OptionFunc {
	return func(o *Option) { o.Retries = n }
}

// WithBackoff sets backoff strategy
func WithBackoff(f BackoffFunc) OptionFunc {
	return func(o *Option) { o.Backoff = f }
}

// WithHeaders adds request headers
func WithHeaders(h map[string]string) OptionFunc {
	return func(o *Option) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		for k, v := range h {
			o.Headers[k] = v
		}
	}
}

// WithForceRetry forces retry for all methods
func WithForceRetry() OptionFunc {
	return func(o *Option) { o.ForceRetry = true }
}

// WithMiddleware adds middleware
func WithMiddleware(mw Middleware) OptionFunc {
	return func(o *Option) { o.Middlewares = append(o.Middlewares, mw) }
}

// WithTimeout sets client timeout
func WithTimeout(timeout time.Duration) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		client.Timeout = timeout
	}
}

// WithMaxIdleConns sets global maximum idle connections
func WithMaxIdleConns(n int) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		if t, ok := client.Transport.(*http.Transport); ok {
			t.MaxIdleConns = n
		}
	}
}

// WithMaxConnsPerHost sets maximum connections per host
func WithMaxConnsPerHost(n int) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		if t, ok := client.Transport.(*http.Transport); ok {
			t.MaxConnsPerHost = n
		}
	}
}

// WithIdleConnTimeout sets idle connection timeout
func WithIdleConnTimeout(d time.Duration) OptionFunc {
	return func(o *Option) {
		client := GetClient()
		if t, ok := client.Transport.(*http.Transport); ok {
			t.IdleConnTimeout = d
		}
	}
}

// HTTPError custom request error type, contains status code, method, URL and response body
type HTTPError struct {
	StatusCode int
	Method     string
	URL        string
	Body       []byte
	Err        error
}

// Error implements error interface
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s failed: status %d, body: %q, error: %v", e.Method, e.URL, e.StatusCode, e.Body, e.Err)
	}
	return fmt.Sprintf("%s %s failed: status %d, body: %q", e.Method, e.URL, e.StatusCode, e.Body)
}

// Request core request function, supports retry, backoff, middleware, buffer pool
// ctx: context, can be used for cancellation or timeout
// method: HTTP method, such as GET, POST, PUT, etc.
// urlStr: request URL
// body: request body content, byte slice
// opts: optional configuration, including retry count, backoff strategy, headers, middleware, etc.
func Request(ctx context.Context, method, urlStr string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	// Initialize default options: 3 retries, default backoff function
	options := &Option{
		Retries: 3,
		Backoff: defaultBackoff,
	}
	for _, opt := range opts {
		opt(options) // Apply user-provided optional configuration
	}

	var lastErr error // Record the last error
	for attempt := 0; attempt <= options.Retries; attempt++ {
		var bodyReader io.Reader
		var buf *bytes.Buffer

		// Optimize request body:
		// Less than 1KB, directly use bytes.NewReader
		// Greater than 1KB, use buffer pool for reuse, reduce memory allocation
		if body != nil {
			if len(body) < 1024 {
				bodyReader = bytes.NewReader(body)
			} else {
				buf = bufferPool.Get().(*bytes.Buffer)
				buf.Reset()
				buf.Write(body)
				bodyReader = buf
			}
		}

		// Create request object, bind context
		req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
		if err != nil {
			if buf != nil {
				bufferPool.Put(buf) // Return buffer to pool on error
			}
			return nil, err
		}

		// Set request headers
		for k, v := range options.Headers {
			req.Header.Set(k, v)
		}

		// Build middleware chain
		doFunc := GetClient().Do // Default HTTP request function
		for i := len(options.Middlewares) - 1; i >= 0; i-- {
			mw := options.Middlewares[i] // Note the closure capture issue
			next := doFunc
			doFunc = func(req *http.Request) (*http.Response, error) {
				return mw(next)(req) // Execute middleware
			}
		}

		// Execute request
		resp, err := doFunc(req)

		// Return buffer to pool immediately after request completion
		if buf != nil {
			bufferPool.Put(buf)
		}

		// If request is successful and status code is 2xx, return directly
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Read response body for error information (max 512 bytes)
		var bodyBytes []byte
		if resp != nil {
			bodyBytes, _ = io.ReadAll(io.LimitReader(resp.Body, 512))
			_ = resp.Body.Close()
		}

		// 4xx errors return directly, no retry
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, &HTTPError{
				StatusCode: resp.StatusCode,
				Method:     method,
				URL:        urlStr,
				Body:       bodyBytes,
			}
		}

		// Record the last error
		lastErr = &HTTPError{
			StatusCode: 0,
			Method:     method,
			URL:        urlStr,
			Body:       bodyBytes,
			Err:        err,
		}

		// Determine if retry is needed
		if attempt < options.Retries {
			// Enable retry for GET/HEAD methods or when ForceRetry is enabled
			if options.ForceRetry || strings.ToUpper(method) == http.MethodGet || strings.ToUpper(method) == http.MethodHead {
				backoff := options.Backoff(attempt) // Calculate backoff time
				select {
				case <-time.After(backoff):
					// Wait for backoff time before retrying
				case <-ctx.Done():
					// Context cancelled or timed out, return
					return nil, ctx.Err()
				}
			} else {
				break // Non-retryable methods exit loop directly
			}
		}
	}

	// All retries failed, return the last error
	return nil, lastErr
}
