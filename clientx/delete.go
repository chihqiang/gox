package clientx

import (
	"context"
	"net/http"
)

// Delete 请求
func Delete(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodDelete, url, body, opts...)
}
