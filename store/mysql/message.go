package mysql

import (
	"github.com/google/uuid"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
	"gorm.io/gorm"
)

// MessageStore 实现MessageStore接口的MySQL实现
type MessageStore struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewMessageStore 创建MySQL消息存储库实例
func NewMessageStore(db *gorm.DB) interfaces.MessageStore {
	return &MessageStore{db: db}
}

// SetLogger 设置日志记录器
func (r *MessageStore) SetLogger(logger *logger.Logger) {
	r.logger = logger
}

// Create 创建消息
func (r *MessageStore) Create(msg *models.Message) error {
	if len(msg.MsgID) == 0 {
		msg.MsgID = uuid.NewString()
	}

	err := r.db.Create(msg).Error
	if err == nil && r.logger != nil {
		r.logger.Info("消息 %s 创建成功", msg.MsgID)
	}
	return err
}

// Update 更新消息
func (r *MessageStore) Update(msg *models.Message) error {
	err := r.db.Save(msg).Error
	if err == nil && r.logger != nil {
		r.logger.Info("消息 %s 更新成功", msg.MsgID)
	}
	return err
}

// Delete 删除消息
func (r *MessageStore) Delete(msgID string) error {
	err := r.db.Where("msg_id = ?", msgID).Delete(&models.Message{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("消息 %s 删除成功", msgID)
	}
	return err
}

// GetByID 根据ID获取消息
func (r *MessageStore) GetByID(msgID string) (*models.Message, error) {
	var msg models.Message
	err := r.db.Where("msg_id = ?", msgID).First(&msg).Error
	if err != nil && r.logger != nil {
		r.logger.Error("获取消息 %s 失败: %v", msgID, err)
		return nil, err
	}
	return &msg, nil
}

// ListByConversation 获取对话的消息列表
func (r *MessageStore) ListByConversation(conversationID string, offset, limit int) ([]*models.Message, error) {
	var msgs []*models.Message
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("order_seq ASC").
		Offset(offset).
		Limit(limit).
		Find(&msgs).Error

	if err == nil && r.logger != nil {
		r.logger.Info("查询到会话 %s 的 %d 条消息记录", conversationID, len(msgs))
	}

	return msgs, err
}

// UpdateStatus 更新消息状态
func (r *MessageStore) UpdateStatus(msgID string, status string) error {
	err := r.db.Model(&models.Message{}).Where("msg_id = ?", msgID).Update("status", status).Error
	if err == nil && r.logger != nil {
		r.logger.Debug("消息 %s 状态更新为 %s", msgID, status)
	}
	return err
}

// UpdateTokenCount 更新消息token数量
func (r *MessageStore) UpdateTokenCount(msgID string, tokenCount int) error {
	err := r.db.Model(&models.Message{}).Where("msg_id = ?", msgID).Update("token_count", tokenCount).Error
	if err == nil && r.logger != nil {
		r.logger.Debug("消息 %s token数量更新为 %d", msgID, tokenCount)
	}
	return err
}

// SetContextEdge 设置消息为上下文边界
func (r *MessageStore) SetContextEdge(msgID string, isContextEdge bool) error {
	err := r.db.Model(&models.Message{}).Where("msg_id = ?", msgID).Update("is_context_edge", isContextEdge).Error
	if err == nil && r.logger != nil {
		r.logger.Debug("消息 %s 上下文边界设置为 %t", msgID, isContextEdge)
	}
	return err
}

// SetVariant 设置消息为变体
func (r *MessageStore) SetVariant(msgID string, isVariant bool) error {
	err := r.db.Model(&models.Message{}).Where("msg_id = ?", msgID).Update("is_variant", isVariant).Error
	if err == nil && r.logger != nil {
		r.logger.Debug("消息 %s 变体设置为 %t", msgID, isVariant)
	}
	return err
}
