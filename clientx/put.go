package clientx

import (
	"context"
	"net/http"
)

// Put 请求
func Put(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodPut, url, body, opts...)
}
