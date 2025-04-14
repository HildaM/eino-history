package interfaces

import (
	"github.com/hildam/eino-history/model"
)

// MessageStore 定义消息存储库接口
type MessageStore interface {
	// Create 创建新消息
	// 参数:
	//   - msg: 要创建的消息对象
	// 返回:
	//   - error: 如果创建过程中发生错误
	Create(msg *models.Message) error

	// Update 更新已有消息
	// 参数:
	//   - msg: 包含更新数据的消息对象
	// 返回:
	//   - error: 如果更新过程中发生错误
	Update(msg *models.Message) error

	// Delete 删除指定ID的消息
	// 参数:
	//   - msgID: 要删除的消息ID
	// 返回:
	//   - error: 如果删除过程中发生错误
	Delete(msgID string) error

	// GetByID 根据ID获取消息
	// 参数:
	//   - msgID: 消息ID
	// 返回:
	//   - *models.Message: 获取到的消息对象
	//   - error: 如果获取过程中发生错误
	GetByID(msgID string) (*models.Message, error)

	// ListByConversation 获取指定会话的消息列表
	// 参数:
	//   - conversationID: 会话ID
	//   - offset: 分页偏移量
	//   - limit: 返回消息数量上限
	// 返回:
	//   - []*models.Message: 消息列表
	//   - error: 如果获取过程中发生错误
	ListByConversation(conversationID string, offset, limit int) ([]*models.Message, error)

	// UpdateStatus 更新消息状态
	// 参数:
	//   - msgID: 消息ID
	//   - status: 新的状态值
	// 返回:
	//   - error: 如果更新过程中发生错误
	UpdateStatus(msgID string, status string) error

	// UpdateTokenCount 更新消息的Token计数
	// 参数:
	//   - msgID: 消息ID
	//   - tokenCount: 新的Token计数
	// 返回:
	//   - error: 如果更新过程中发生错误
	UpdateTokenCount(msgID string, tokenCount int) error

	// SetContextEdge 设置消息是否为上下文边界
	// 参数:
	//   - msgID: 消息ID
	//   - isContextEdge: 是否为上下文边界
	// 返回:
	//   - error: 如果设置过程中发生错误
	SetContextEdge(msgID string, isContextEdge bool) error

	// SetVariant 设置消息是否为变体
	// 参数:
	//   - msgID: 消息ID
	//   - isVariant: 是否为变体
	// 返回:
	//   - error: 如果设置过程中发生错误
	SetVariant(msgID string, isVariant bool) error
}

// ConversationStore 定义对话存储库接口
type ConversationStore interface {
	// Create 创建新会话
	// 参数:
	//   - conv: 要创建的会话对象
	// 返回:
	//   - error: 如果创建过程中发生错误
	Create(conv *models.Conversation) error

	// Update 更新已有会话
	// 参数:
	//   - conv: 包含更新数据的会话对象
	// 返回:
	//   - error: 如果更新过程中发生错误
	Update(conv *models.Conversation) error

	// Delete 删除指定ID的会话
	// 参数:
	//   - convID: 要删除的会话ID
	// 返回:
	//   - error: 如果删除过程中发生错误
	Delete(convID string) error

	// GetByID 根据ID获取会话
	// 参数:
	//   - convID: 会话ID
	// 返回:
	//   - *models.Conversation: 获取到的会话对象
	//   - error: 如果获取过程中发生错误
	GetByID(convID string) (*models.Conversation, error)

	// FirstOrCreat 根据ID查找会话，如不存在则创建
	// 参数:
	//   - convID: 会话ID
	// 返回:
	//   - *models.Conversation: 查找到或新创建的会话对象
	//   - error: 如果操作过程中发生错误
	FirstOrCreat(convID string) (*models.Conversation, error)

	// List 获取会话列表
	// 参数:
	//   - offset: 分页偏移量
	//   - limit: 返回会话数量上限
	// 返回:
	//   - []*models.Conversation: 会话列表
	//   - error: 如果获取过程中发生错误
	List(offset, limit int) ([]*models.Conversation, error)

	// Archive 归档指定会话
	// 参数:
	//   - convID: 要归档的会话ID
	// 返回:
	//   - error: 如果归档过程中发生错误
	Archive(convID string) error

	// Unarchive 取消归档指定会话
	// 参数:
	//   - convID: 要取消归档的会话ID
	// 返回:
	//   - error: 如果取消归档过程中发生错误
	Unarchive(convID string) error

	// Pin 置顶指定会话
	// 参数:
	//   - convID: 要置顶的会话ID
	// 返回:
	//   - error: 如果置顶过程中发生错误
	Pin(convID string) error

	// Unpin 取消置顶指定会话
	// 参数:
	//   - convID: 要取消置顶的会话ID
	// 返回:
	//   - error: 如果取消置顶过程中发生错误
	Unpin(convID string) error
}

// AttachmentStore 定义附件存储库接口
type AttachmentStore interface {
	// Create 创建新附件
	// 参数:
	//   - attachment: 要创建的附件对象
	// 返回:
	//   - error: 如果创建过程中发生错误
	Create(attachment *models.Attachment) error

	// Update 更新已有附件
	// 参数:
	//   - attachment: 包含更新数据的附件对象
	// 返回:
	//   - error: 如果更新过程中发生错误
	Update(attachment *models.Attachment) error

	// Delete 删除指定ID的附件
	// 参数:
	//   - attachID: 要删除的附件ID
	// 返回:
	//   - error: 如果删除过程中发生错误
	Delete(attachID string) error

	// GetByID 根据ID获取附件
	// 参数:
	//   - attachID: 附件ID
	// 返回:
	//   - *models.Attachment: 获取到的附件对象
	//   - error: 如果获取过程中发生错误
	GetByID(attachID string) (*models.Attachment, error)

	// ListByMessage 获取指定消息的附件列表
	// 参数:
	//   - messageID: 消息ID
	// 返回:
	//   - []*models.Attachment: 附件列表
	//   - error: 如果获取过程中发生错误
	ListByMessage(messageID string) ([]*models.Attachment, error)
}

// MessageAttachmentStore 定义消息-附件关联存储库接口
type MessageAttachmentStore interface {
	// Create 创建消息与附件的关联
	// 参数:
	//   - messageAttachment: 要创建的消息附件关联对象
	// 返回:
	//   - error: 如果创建过程中发生错误
	Create(messageAttachment *models.MessageAttachment) error

	// Delete 根据ID删除关联
	// 参数:
	//   - id: 关联记录的ID
	// 返回:
	//   - error: 如果删除过程中发生错误
	Delete(id uint64) error

	// ListByMessage 获取指定消息的所有附件关联
	// 参数:
	//   - messageID: 消息ID
	// 返回:
	//   - []*models.MessageAttachment: 消息附件关联列表
	//   - error: 如果获取过程中发生错误
	ListByMessage(messageID string) ([]*models.MessageAttachment, error)

	// ListByAttachment 获取指定附件的所有消息关联
	// 参数:
	//   - attachmentID: 附件ID
	// 返回:
	//   - []*models.MessageAttachment: 消息附件关联列表
	//   - error: 如果获取过程中发生错误
	ListByAttachment(attachmentID string) ([]*models.MessageAttachment, error)
}
