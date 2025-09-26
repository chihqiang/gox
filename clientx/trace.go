package clientx

import (
	"context"
	"net/http"
)

// Trace 请求
func Trace(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodTrace, url, body, opts...)
}
