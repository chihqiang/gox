package clientx

import (
	"context"
	"net/http"
)

// Head request
func Head(ctx context.Context, url string, opts ...OptionFunc) (*http.Response, error) {
	return Request(ctx, http.MethodHead, url, nil, opts...)
}
