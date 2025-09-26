package logx

import (
	"io"
	"os"
	"sync"
)

var (
	std     *Logger   // 全局默认 Logger 实例
	stdOnce sync.Once // 确保全局 Logger 只初始化一次（线程安全）
)

// _std 返回全局 Logger 实例（单例模式）
// 第一次调用时会初始化 Logger，并设置 callDepth
func _std() *Logger {
	stdOnce.Do(func() {
		// 创建一个 Logger，输出到标准错误
		_log := New(os.Stderr)
		_log.callDepth = 3 // 调用深度偏移，用于正确显示文件/行号
		std = _log
	})
	return std
}

// SetOutput 设置全局 Logger 的输出目标（线程安全）
func SetOutput(w io.Writer) {
	_std().SetOutput(w)
}

// SetPrefix 设置全局 Logger 的日志前缀（线程安全）
func SetPrefix(p string) {
	_std().SetPrefix(p)
}

// SetFormatter 设置全局 Logger 的日志格式化函数（线程安全）
func SetFormatter(fn Formatter) {
	_std().SetFormatter(fn)
}

// Debug 记录 Debug 级别日志
func Debug(format string, v ...any) {
	_std().Debug(format, v...)
}

// Info 记录 Info 级别日志
func Info(format string, v ...any) {
	_std().Info(format, v...)
}

// Warn 记录 Warn 级别日志
func Warn(format string, v ...any) {
	_std().Warn(format, v...)
}

// Error 记录 Error 级别日志
func Error(format string, v ...any) {
	_std().Error(format, v...)
}

// Log 记录指定 Level 级别的日志
// 如果等级高于 Logger 设置的最低等级，则日志会被输出
func Log(level Level, format string, v ...any) error {
	return _std().Log(level, format, v...)
}
