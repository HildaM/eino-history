package eino

import (
	"github.com/cloudwego/eino/schema"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/interfaces"
	"github.com/hildam/eino-history/store/provider"
)

// History 是聊天历史管理的主要结构体
type History struct {
	mr         interfaces.MessageStore
	cr         interfaces.ConversationStore
	dbProvider provider.Provider // 持有数据库提供者实例
}

// NewDefaultEinoHistory 创建一个使用MySQL作为默认存储的历史实例
// 参数:
//   - dsn: 数据库连接字符串
//   - loglevel: 日志级别("error", "info", "debug")
//
// 返回:
//   - *History: 新创建的历史实例
func NewDefaultEinoHistory(dsn, loglevel string) *History {
	// 如果日志级别为 info 以上，debug 为 true，否则是 false
	debug := false
	if loglevel == "debug" {
		debug = true
	}

	config := &provider.Config{
		DSN:      dsn,
		Type:     provider.TypeMySQL, // 明确指定MySQL
		Debug:    debug,              // 默认关闭调试
		LogLevel: loglevel,
	}

	dbProvider, err := provider.CreateProvider(config)
	if err != nil {
		panic(err)
	}

	return &History{
		mr:         dbProvider.GetMessageStore(),
		cr:         dbProvider.GetConversationStore(),
		dbProvider: dbProvider,
	}
}

// NewEinoHistoryWithProvider 创建一个使用指定数据库提供者的历史实例
// 参数:
//   - dsn: 数据库连接字符串
//   - dbType: 数据库类型(MySQL或Redis)
//   - debug: 是否启用调试
//   - logLevel: 日志级别("error", "info", "debug")
//
// 返回:
//   - *History: 新创建的历史实例
func NewEinoHistoryWithProvider(dsn string, dbType provider.Type, debug bool, logLevel string) *History {
	config := &provider.Config{
		DSN:      dsn,
		Type:     dbType,
		Debug:    debug,
		LogLevel: logLevel,
	}

	dbProvider, err := provider.CreateProvider(config)
	if err != nil {
		panic(err)
	}

	return &History{
		mr:         dbProvider.GetMessageStore(),
		cr:         dbProvider.GetConversationStore(),
		dbProvider: dbProvider,
	}
}

// Close 关闭数据库连接
// 返回:
//   - error: 如果关闭过程中发生错误
func (x *History) Close() error {
	if x.dbProvider != nil {
		return x.dbProvider.Close()
	}
	return nil
}

// SaveMessage 存储消息
// 参数:
//   - mess: 要存储的消息
//   - convID: 会话ID
//
// 返回:
//   - error: 如果存储过程中发生错误
func (x *History) SaveMessage(mess *schema.Message, convID string) error {
	return x.mr.Create(&models.Message{
		Role:           string(mess.Role),
		Content:        mess.Content,
		ConversationID: convID,
	})
}

// GetHistory 根据会话ID获取聊天历史
// 参数:
//   - convID: 会话ID
//   - limit: 返回的消息数量上限，0表示使用默认值(100)
//
// 返回:
//   - []*schema.Message: 消息列表
//   - error: 如果获取过程中发生错误
func (x *History) GetHistory(convID string, limit int) (list []*schema.Message, err error) {
	if limit == 0 {
		limit = 100
	}
	// 如果convID数据不存在，则创建
	_, err = x.cr.FirstOrCreat(convID)
	if err != nil {
		return
	}
	// 最多取100条
	mess, err := x.mr.ListByConversation(convID, 0, limit)
	if err != nil {
		return
	}
	list = messageList2ChatHistory(mess)
	return
}
