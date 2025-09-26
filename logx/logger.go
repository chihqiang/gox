package logx

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// New 创建一个 Logger 实例
// 参数 w 指定日志输出目标（可以是 os.Stdout、os.Stderr 或文件等）
func New(w io.Writer) *Logger {
	l := &Logger{}
	l.SetOutput(w)
	l.SetFormatter(DefaultFormatter) // 使用默认格式化函数
	return l
}

// Logger 日志对象
type Logger struct {
	mu        sync.RWMutex // 读写锁，保证并发安全
	writer    io.Writer    // 日志输出目标
	prefix    string       // 日志前缀
	formatter Formatter    // 日志格式化函数
	callDepth int          // runtime.Caller 层级偏移，用于正确显示调用文件和行号
}

// SetOutput 设置日志输出目标（线程安全）
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = w
}

// SetPrefix 设置日志前缀（线程安全）
func (l *Logger) SetPrefix(p string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = p
}

// SetFormatter 设置日志格式化函数（线程安全）
func (l *Logger) SetFormatter(fn Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = fn
}

// Debug 输出 Debug 级别日志
func (l *Logger) Debug(format string, v ...any) {
	_ = l.Log(LevelDebug, format, v...)
}

// Info 输出 Info 级别日志
func (l *Logger) Info(format string, v ...any) {
	_ = l.Log(LevelInfo, format, v...)
}

// Warn 输出 Warn 级别日志
func (l *Logger) Warn(format string, v ...any) {
	_ = l.Log(LevelWarn, format, v...)
}

// Error 输出 Error 级别日志
func (l *Logger) Error(format string, v ...any) {
	_ = l.Log(LevelError, format, v...)
}

// Log 输出指定等级的日志
// 1. 根据 callDepth 获取调用文件和行号
// 2. 使用 Formatter 格式化日志条目
// 3. 写入日志输出目标（writer），如果 writer 为 nil，默认使用 os.Stdout
func (l *Logger) Log(level Level, format string, v ...any) error {
	// 并发安全读取 Logger 当前状态
	l.mu.RLock()
	prefix := l.prefix
	formatter := l.formatter
	writer := l.writer
	callDepth := l.callDepth
	if callDepth == 0 {
		callDepth = 2 // 默认偏移，保证 runtime.Caller 获取到正确的调用者位置
	}
	l.mu.RUnlock()

	// 格式化日志内容
	msg := fmt.Sprintf(format, v...)

	// 获取调用文件和行号
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???" // 无法获取时使用占位符
		line = 0
	}

	// 构造日志条目并格式化
	entry := LogEntry{
		Time:      time.Now(),
		Level:     level,
		Prefix:    prefix,
		CallDepth: callDepth,
		File:      file,
		Line:      line,
		Message:   msg,
	}
	logStr := formatter(entry)

	// 输出日志，如果 writer 为 nil，则默认输出到 stdout
	if writer == nil {
		writer = os.Stdout
	}
	_, err := writer.Write([]byte(logStr + "\n"))
	return err
}
