package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	// ErrorLevel 仅记录错误
	ErrorLevel LogLevel = iota
	// InfoLevel 记录错误和信息
	InfoLevel
	// DebugLevel 记录错误、信息和调试信息
	DebugLevel
)

// 字符串常量，用于表示日志级别
const (
	// LevelError 错误级别的字符串表示
	LevelError = "error"
	// LevelInfo 信息级别的字符串表示
	LevelInfo = "info"
	// LevelDebug 调试级别的字符串表示
	LevelDebug = "debug"
)

// 默认值
const (
	// DefaultLogLevel 默认的日志级别
	DefaultLogLevel = LevelInfo
)

// LogLevelNames 日志级别名称映射
var LogLevelNames = map[string]LogLevel{
	LevelError: ErrorLevel,
	LevelInfo:  InfoLevel,
	LevelDebug: DebugLevel,
}

// Logger 提供不同级别的日志记录功能
type Logger struct {
	// Enabled 控制是否启用日志记录
	Enabled bool
	// Level 日志级别
	Level LogLevel
	// Prefix 日志前缀
	Prefix string
	// logger 内部logger实例
	logger *log.Logger
}

// NewWithLevelName 使用字符串日志级别名称创建日志记录器
// 参数:
//   - enabled: 是否启用日志
//   - levelName: 日志级别名称 ("error", "info", "debug")
//   - prefix: 日志前缀
//
// 返回:
//   - *Logger: 新创建的日志记录器
func NewWithLevelName(enabled bool, levelName string, prefix string) *Logger {
	level, ok := LogLevelNames[levelName]
	if !ok {
		// 如果日志级别无效，使用错误级别而不是默认设置成info
		// 这样可以让上层调用方知道他们的配置有问题
		return &Logger{
			Enabled: enabled,
			Level:   ErrorLevel,
			Prefix:  prefix,
			logger:  log.New(os.Stdout, "", log.LstdFlags),
		}
	}
	return NewWithLevel(enabled, level, prefix)
}

// NewWithIntLevel 使用整数日志级别创建日志记录器
// 参数:
//   - enabled: 是否启用日志
//   - levelInt: 整数日志级别 (0=Error, 1=Info, 2=Debug)
//   - prefix: 日志前缀
//
// 返回:
//   - *Logger: 新创建的日志记录器
func NewWithIntLevel(enabled bool, levelInt int, prefix string) *Logger {
	var level LogLevel
	switch levelInt {
	case 0:
		level = ErrorLevel
	case 1:
		level = InfoLevel
	case 2:
		level = DebugLevel
	default:
		// 如果日志级别无效，使用错误级别
		level = ErrorLevel
	}
	return NewWithLevel(enabled, level, prefix)
}

// NewWithLevel 使用LogLevel创建日志记录器
// 参数:
//   - enabled: 是否启用日志
//   - level: 日志级别枚举值
//   - prefix: 日志前缀
//
// 返回:
//   - *Logger: 新创建的日志记录器
func NewWithLevel(enabled bool, level LogLevel, prefix string) *Logger {
	return &Logger{
		Enabled: enabled,
		Level:   level,
		Prefix:  prefix,
		logger:  log.New(os.Stdout, "", log.LstdFlags),
	}
}

// formatMessage 格式化日志消息，添加时间戳、前缀和级别
// 参数:
//   - level: 日志级别字符串
//   - format: 格式化字符串
//   - args: 格式化参数
//
// 返回:
//   - string: 格式化后的日志消息
func (l *Logger) formatMessage(level, format string, args ...interface{}) string {
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s-%s] %s", l.Prefix, level, message)
}

// Error 记录错误级别的日志
// 参数:
//   - format: 格式化字符串
//   - args: 格式化参数
func (l *Logger) Error(format string, args ...interface{}) {
	if !l.Enabled {
		return
	}
	l.logger.Print(l.formatMessage("ERROR", format, args...))
}

// Info 记录信息级别的日志
// 参数:
//   - format: 格式化字符串
//   - args: 格式化参数
func (l *Logger) Info(format string, args ...interface{}) {
	if !l.Enabled || l.Level < InfoLevel {
		return
	}
	l.logger.Print(l.formatMessage("INFO", format, args...))
}

// Debug 记录调试级别的日志
// 参数:
//   - format: 格式化字符串
//   - args: 格式化参数
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.Enabled || l.Level < DebugLevel {
		return
	}
	l.logger.Print(l.formatMessage("DEBUG", format, args...))
}

// WithPrefix 创建一个带有特定前缀的新logger实例
// 参数:
//   - prefix: 新的日志前缀
//
// 返回:
//   - *Logger: 新创建的日志记录器，继承原有logger的配置，但使用新前缀
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		Enabled: l.Enabled,
		Level:   l.Level,
		Prefix:  prefix,
		logger:  l.logger,
	}
}

// SetOutput 设置日志输出目标
// 参数:
//   - w: 输出的文件对象
func (l *Logger) SetOutput(w *os.File) {
	l.logger.SetOutput(w)
}

// LogFunc 返回一个记录执行时间的函数包装器
// 参数:
//   - name: 被包装函数的名称
//
// 返回:
//   - func(): 一个在执行结束时记录耗时的函数
func (l *Logger) LogFunc(name string) func() {
	start := time.Now()
	l.Debug("开始执行 %s", name)
	return func() {
		l.Debug("完成执行 %s, 耗时: %v", name, time.Since(start))
	}
}
