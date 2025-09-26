package httpx

import "fmt"

// NewCodeMsg 构造函数
func NewCodeMsg(code int, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}

// CodeMsg 自定义业务错误
type CodeMsg struct {
	Code int
	Msg  string
}

// 实现 error 接口
func (e *CodeMsg) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Msg)
}
