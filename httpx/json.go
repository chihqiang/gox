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
	// 从缓冲池获取缓冲区
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jsonBufferPool.Put(buf)
	// 使用缓冲区进行编码
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON response: %w", err)
	}
	// 将编码后的内容写入响应
	_, err := w.Write(buf.Bytes())
	return err
}
