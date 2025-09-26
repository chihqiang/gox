package clientx

import (
	"context"
	"net/http"
)

// Head 请求
func Head(ctx context.Context, url string, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodHead, url, nil, opts...)
}
