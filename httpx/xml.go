package httpx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"
)

var xmlBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func XML(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(status)
	// 从缓冲池获取缓冲区
	buf := xmlBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer xmlBufferPool.Put(buf)
	// 使用缓冲区进行编码
	if err := xml.NewEncoder(buf).Encode(v); err != nil {
		return fmt.Errorf("failed to encode XML response: %w", err)
	}
	// 将编码后的内容写入响应
	_, err := w.Write(buf.Bytes())
	return err
}

// XmlResponse writes v into w with http.StatusOK.
func XmlResponse[T any](w http.ResponseWriter, v T) error {
	return XML(w, http.StatusOK, wrapXmlBaseResponse[T](v))
}

func wrapXmlBaseResponse[T any](v T) BaseXmlResponse[T] {
	base := wrapBaseResponse(v)
	return BaseXmlResponse[T]{
		Version:      xmlVersion,
		Encoding:     xmlEncoding,
		BaseResponse: base,
	}
}
