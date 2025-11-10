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
	// Get buffer from buffer pool
	buf := xmlBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer xmlBufferPool.Put(buf)
	// Use buffer for encoding
	if err := xml.NewEncoder(buf).Encode(v); err != nil {
		return fmt.Errorf("failed to encode XML response: %w", err)
	}
	// Write encoded content to response
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
