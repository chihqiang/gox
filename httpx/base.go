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
	// 自定义业务错误
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

// Unauthorized 未登录/未认证返回
func Unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized) // 401
}

// Forbidden 权限不足返回
func Forbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden) // 403
}

// BadRequest 400 错误返回
func BadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

// NotFound 404 错误返回
func NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

// InternalServerError 500 错误返回
func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

// Status 写入指定的HTTP状态码
func Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}
