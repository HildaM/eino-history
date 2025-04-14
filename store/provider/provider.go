package provider

import (
	"github.com/hildam/eino-history/store/interfaces"
)

// Type 数据库类型
type Type string

const (
	// TypeMySQL MySQL数据库类型
	TypeMySQL Type = "mysql"
	// TypeRedis Redis数据库类型
	TypeRedis Type = "redis"
)

// Provider 定义数据库提供者接口
type Provider interface {
	// GetMessageStore 获取消息存储库
	GetMessageStore() interfaces.MessageStore
	// GetConversationStore 获取对话存储库
	GetConversationStore() interfaces.ConversationStore
	// GetAttachmentStore 获取附件存储库
	GetAttachmentStore() interfaces.AttachmentStore
	// GetMessageAttachmentStore 获取消息附件关联存储库
	GetMessageAttachmentStore() interfaces.MessageAttachmentStore
	// Close 关闭数据库连接
	Close() error
}

// Config 数据库配置
type Config struct {
	// DSN 数据库连接字符串
	DSN string
	// Type 数据库类型
	Type Type
	// Debug 是否开启调试模式
	Debug bool
	// LogLevel 日志级别
	LogLevel string
}
