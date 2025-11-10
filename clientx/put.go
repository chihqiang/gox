package clientx

import (
	"context"
	"net/http"
)

// Put request
func Put(ctx context.Context, url string, body []byte, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodPut, url, body, opts...)
}
