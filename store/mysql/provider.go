package mysql

import (
	"fmt"
	"time"

	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Provider 实现Provider接口的MySQL实现
type Provider struct {
	db                    *gorm.DB
	messageRepo           interfaces.MessageStore
	conversationRepo      interfaces.ConversationStore
	attachmentRepo        interfaces.AttachmentStore
	messageAttachmentRepo interfaces.MessageAttachmentStore
	logger                *logger.Logger
}

// NewProvider 创建MySQL提供者实例
// 参数:
//   - dsn: 数据库连接字符串
//   - loggingEnabled: 是否启用日志
//   - logLevel: 日志级别("error", "info", "debug")
//
// 返回:
//   - *Provider: 新创建的MySQL提供者
//   - error: 如果创建过程中发生错误
func NewProvider(dsn string, loggingEnabled bool, logLevel string) (*Provider, error) {
	// 使用更明确的工厂方法创建日志记录器
	customLogger := logger.NewWithLevelName(loggingEnabled, logLevel, "MySQL")
	customLogger.Info("初始化MySQL Provider...")

	// 设置GORM日志级别
	var gormLogLevel gormlogger.LogLevel
	if loggingEnabled {
		gormLogLevel = gormlogger.Info
	} else {
		gormLogLevel = gormlogger.Error // 只记录错误
	}

	// 配置GORM
	config := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %v", err)
	}

	// 获取数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接池失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移数据库表结构
	if err = autoMigrateTables(db); err != nil {
		return nil, fmt.Errorf("自动迁移数据库表结构失败: %v", err)
	}

	customLogger.Info("数据库表结构自动迁移完成")

	// 创建提供者实例
	provider := &Provider{
		db:     db,
		logger: customLogger,
	}

	// 初始化仓库
	provider.messageRepo = NewMessageStore(db)
	provider.conversationRepo = NewConversationStore(db)
	provider.attachmentRepo = NewAttachmentStore(db)
	provider.messageAttachmentRepo = NewMessageAttachmentStore(db)

	// 注入日志记录器到仓库中
	setLoggers(provider)

	customLogger.Info("MySQL存储初始化成功")
	return provider, nil
}

// setLoggers 将日志记录器注入到各个仓库中
// 参数:
//   - p: Provider实例，包含需要设置logger的仓库
func setLoggers(p *Provider) {
	if messageRepo, ok := p.messageRepo.(*MessageStore); ok {
		messageRepo.SetLogger(p.logger)
	}

	if conversationRepo, ok := p.conversationRepo.(*ConversationStore); ok {
		conversationRepo.SetLogger(p.logger)
	}

	if attachmentRepo, ok := p.attachmentRepo.(*AttachmentStore); ok {
		attachmentRepo.SetLogger(p.logger)
	}

	if messageAttachmentRepo, ok := p.messageAttachmentRepo.(*MessageAttachmentStore); ok {
		messageAttachmentRepo.SetLogger(p.logger)
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
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	p.logger.Info("关闭MySQL连接")
	return sqlDB.Close()
}

// GetDB 获取gorm.DB实例（用于兼容旧代码）
// 返回:
//   - *gorm.DB: 底层的gorm.DB实例
func (p *Provider) GetDB() *gorm.DB {
	return p.db
}

// autoMigrateTables 自动迁移数据库表结构
// 参数:
//   - db: gorm.DB实例
//
// 返回:
//   - error: 如果迁移过程中发生错误
func autoMigrateTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
		&models.Attachment{},
		&models.MessageAttachment{},
	)
}
