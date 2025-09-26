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
		return fmt.Sprintf("%s%+d", base, val) // %+d 保留符号
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

// Color 返回日志等级对应的彩色输出（使用 github.com/fatih/color）
func (l Level) Color() *color.Color {
	switch {
	case l >= LevelError: // 错误
		return color.New(color.FgHiRed, color.Bold) // 亮红色+粗体
	case l >= LevelWarn: // 警告
		return color.New(color.FgYellow, color.Bold) // 粗黄色
	case l >= LevelInfo: // 信息
		return color.New(color.FgGreen) // 绿色
	case l >= LevelDebug: // 调试
		return color.New(color.FgBlue) // 蓝色
	default:
		return color.New(color.FgWhite) // 默认白色
	}
}
