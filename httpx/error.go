package httpx

import "fmt"

// NewCodeMsg constructor
func NewCodeMsg(code int, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}

// CodeMsg custom business error
type CodeMsg struct {
	Code int
	Msg  string
}

// Implement error interface
func (e *CodeMsg) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Msg)
}
