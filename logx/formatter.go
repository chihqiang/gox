package logx

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

// LogEntry 日志条目结构体
type LogEntry struct {
	Time      time.Time `json:"time" xml:"time"`       // 日志发生的时间
	Level     Level     `json:"level" xml:"level"`     // 日志等级（如 TRACE、INFO、ERROR 等）
	Prefix    string    `json:"prefix" xml:"prefix"`   // 日志前缀，用于区分模块或子系统，可为空
	File      string    `json:"file" xml:"file"`       // 日志所在文件路径（相对路径或经过格式化的路径）
	Line      int       `json:"line" xml:"line"`       // 日志所在文件的行号
	Message   string    `json:"message" xml:"message"` // 日志内容正文
	CallDepth int       `json:"-" xml:"-"`             // 堆栈深度，用于获取调用源位置（文件和行号）
}

// Formatter 格式化函数类型
// 输入: 日志条目
// 输出: 格式化后的日志字符串
type Formatter func(entry LogEntry) []byte

var DefaultFormatter Formatter = func(entry LogEntry) []byte {
	// 时间格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	// 日志等级大写
	level := entry.Level.String()

	fileLine := fmt.Sprintf("[%s:%d]", TrimCallerPath(entry.File, 1), entry.Line)
	// 日志前缀
	prefix := ""
	if entry.Prefix != "" {
		prefix = entry.Prefix + ": "
	}
	// 自定义默认输出格式
	logStr := fmt.Sprintf("%s %s %s %s%s",
		timestamp,
		entry.Level.Color().Sprint(level),
		color.New(color.FgHiBlack).Sprint(fileLine),
		color.New(color.FgHiBlack).Add(color.Bold).Sprint(prefix),
		entry.Level.Color().Sprint(entry.Message),
	)
	return []byte(logStr + "\n")
}

func TrimCallerPath(path string, n int) string {
	// lovely borrowed from zap
	// nb. To make sure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.
	// Return the full path if n is 0.
	if n <= 0 {
		return path
	}
	// Find the last separator.
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return path
	}
	for i := 0; i < n-1; i++ {
		// Find the penultimate separator.
		idx = strings.LastIndexByte(path[:idx], '/')
		if idx == -1 {
			return path
		}
	}
	return path[idx+1:]
}
