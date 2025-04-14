package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
)

// Provider 实现Provider接口的Redis实现
type Provider struct {
	client                *redis.Client
	messageRepo           interfaces.MessageStore
	conversationRepo      interfaces.ConversationStore
	attachmentRepo        interfaces.AttachmentStore
	messageAttachmentRepo interfaces.MessageAttachmentStore
	logger                *logger.Logger
}

// NewProvider 创建Redis提供者实例
// 参数:
//   - dsn: Redis连接字符串
//   - loggingEnabled: 是否启用日志
//   - logLevel: 日志级别("error", "info", "debug")
//
// 返回:
//   - *Provider: 新创建的Redis提供者
//   - error: 如果创建过程中发生错误
func NewProvider(dsn string, loggingEnabled bool, logLevel string) (*Provider, error) {
	// 使用更明确的工厂方法创建日志记录器
	customLogger := logger.NewWithLevelName(loggingEnabled, logLevel, "Redis")
	customLogger.Info("初始化连接 %s", dsn)

	// 解析Redis URL
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析Redis URL失败: %v", err)
	}

	// 创建Redis客户端
	client := redis.NewClient(opt)
	ctx := context.Background()

	// 测试连接
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	customLogger.Info("连接成功")

	// 创建提供者实例
	provider := &Provider{
		client: client,
		logger: customLogger,
	}

	// 为了兼容现有代码，我们需要先创建仓库，然后再设置logger
	debug := loggingEnabled
	provider.messageRepo = NewMessageStore(client, debug)
	provider.conversationRepo = NewConversationStore(client, debug)
	provider.attachmentRepo = NewAttachmentStore(client, debug)
	provider.messageAttachmentRepo = NewMessageAttachmentStore(client, debug)

	// 设置日志记录器
	setLoggers(provider)

	return provider, nil
}

// setLoggers 将日志记录器注入到各个仓库中
// 参数:
//   - p: Provider实例，包含需要设置logger的仓库
func setLoggers(p *Provider) {
	// 为所有支持SetLogger方法的仓库设置logger
	if messageRepo, ok := p.messageRepo.(*MessageStore); ok && messageRepo != nil {
		messageRepo.SetLogger(p.logger)
	}
}

// GetMessageStore 获取消息存储库
// 返回:
//   - interfaces.MessageStore: 消息存储库实例
func (p *Provider) GetMessageStore() interfaces.MessageStore {
	return p.messageRepo
}

// GetConversationStore 获取对话存储库
// 返回:
//   - interfaces.ConversationStore: 对话存储库实例
func (p *Provider) GetConversationStore() interfaces.ConversationStore {
	return p.conversationRepo
}

// GetAttachmentStore 获取附件存储库
// 返回:
//   - interfaces.AttachmentStore: 附件存储库实例
func (p *Provider) GetAttachmentStore() interfaces.AttachmentStore {
	return p.attachmentRepo
}

// GetMessageAttachmentStore 获取消息附件关联存储库
// 返回:
//   - interfaces.MessageAttachmentStore: 消息附件关联存储库实例
func (p *Provider) GetMessageAttachmentStore() interfaces.MessageAttachmentStore {
	return p.messageAttachmentRepo
}

// Close 关闭数据库连接
// 返回:
//   - error: 如果关闭过程中发生错误
func (p *Provider) Close() error {
	p.logger.Info("关闭Redis连接")
	return p.client.Close()
}
