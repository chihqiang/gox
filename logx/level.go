package logx

import (
	"fmt"
	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = -4 // 调试级别
	LevelInfo  Level = 0  // 普通信息
	LevelWarn  Level = 4  // 警告
	LevelError Level = 8  // 错误
)

// String 返回日志级别对应的字符串表示
// 如果是非标准等级，会在基础等级后加上偏移量，例如 DEBUG+1
func (l Level) String() string {
	str := func(base string, val Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l < LevelInfo:
		return str("DEBUG", l-LevelDebug)
	case l < LevelWarn:
		return str("INFO", l-LevelInfo)
	case l < LevelError:
		return str("WARN", l-LevelWarn)
	default:
		return str("ERROR", l-LevelError)
	}
}

// MarshalJSON 实现了 json.Marshaler 接口
// 当 LogEntry 被 json.Marshal 序列化时，会调用这个方法
// 目的是将 Level 类型序列化为对应的字符串（如 "INFO", "ERROR"）而不是整数
func (l Level) MarshalJSON() ([]byte, error) {
	// 调用 l.String() 获取 Level 对应的字符串
	// 然后加上双引号，返回字节切片以符合 JSON 字符串格式
	return []byte(`"` + l.String() + `"`), nil
}

// Color 返回日志等级对应的彩色输出（使用 github.com/fatih/color）
func (l Level) Color() *color.Color {
	switch {
	case l >= LevelError: // 8 及以上
		return color.New(color.FgHiRed, color.Bold)
	case l >= LevelWarn: // 4 及以上
		return color.New(color.FgYellow, color.Bold)
	case l >= LevelInfo: // 0 及以上
		return color.New(color.FgGreen)
	case l >= LevelDebug: // -4 及以上
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}
