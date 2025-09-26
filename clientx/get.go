package clientx

import (
	"context"
	"net/http"
)

// Get 请求
func Get(ctx context.Context, url string, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodGet, url, nil, opts...)
}
