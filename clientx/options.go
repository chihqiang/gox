package clientx

import (
	"context"
	"net/http"
)

// Options 请求
func Options(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodOptions, url, body, opts...)
}
