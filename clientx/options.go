package clientx

import (
	"context"
	"net/http"
)

// Options request
func Options(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodOptions, url, body, opts...)
}
