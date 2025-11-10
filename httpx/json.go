package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// JsonResponse writes v into w with http.StatusOK.
func JsonResponse[T any](w http.ResponseWriter, v T) error {
	return JSON(w, http.StatusOK, wrapBaseResponse[T](v))
}

// JSON writes v into w with 200 OK.
func JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Get buffer from buffer pool
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jsonBufferPool.Put(buf)
	// Use buffer for encoding
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON response: %w", err)
	}
	// Write encoded content to response
	_, err := w.Write(buf.Bytes())
	return err
}
