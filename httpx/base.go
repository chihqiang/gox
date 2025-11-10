package httpx

import (
	"encoding/xml"
	"net/http"
)

const (
	xmlVersion  = "1.0"
	xmlEncoding = "UTF-8"

	// BusinessCodeOK represents the business code for success.
	BusinessCodeOK = 0
	// BusinessMsgOk represents the business message for success.
	BusinessMsgOk = "ok"
	// BusinessCodeError represents the business code for error.
	BusinessCodeError = -1
)

// BaseResponse is the base response struct.
type BaseResponse[T any] struct {
	// Code represents the business code, not the http status code.
	Code int `json:"code" xml:"code"`
	// Msg represents the business message, if Code = BusinessCodeOK,
	// and Msg is empty, then the Msg will be set to BusinessMsgOk.
	Msg string `json:"msg" xml:"msg"`
	// Data represents the business data.
	Data T `json:"data,omitempty" xml:"data,omitempty"`
}
type BaseXmlResponse[T any] struct {
	XMLName  xml.Name `xml:"xml"`
	Version  string   `xml:"version,attr"`
	Encoding string   `xml:"encoding,attr"`
	BaseResponse[T]
}

func wrapBaseResponse[T any](v T) BaseResponse[T] {
	var resp BaseResponse[T]
	switch data := any(v).(type) {
	// Custom business error
	case *CodeMsg:
		resp.Code = data.Code
		resp.Msg = data.Msg
	case CodeMsg:
		resp.Code = data.Code
		resp.Msg = data.Msg
	case error:
		resp.Code = BusinessCodeError
		resp.Msg = data.Error()
	default:
		resp.Code = BusinessCodeOK
		resp.Msg = BusinessMsgOk
		resp.Data = v
	}
	return resp
}

// Ok writes HTTP 200 OK into w.
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// Unauthorized returns 401 unauthorized response
func Unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized) // 401
}

// Forbidden returns 403 forbidden response
func Forbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden) // 403
}

// BadRequest returns 400 bad request response
func BadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

// NotFound returns 404 not found response
func NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

// InternalServerError returns 500 internal server error response
func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

// Status writes the specified HTTP status code
func Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}
