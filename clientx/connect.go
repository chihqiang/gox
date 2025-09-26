package clientx

import (
	"context"
	"net/http"
)

// Connect 请求
func Connect(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodConnect, url, body, opts...)
}
