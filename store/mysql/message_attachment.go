package mysql

import (
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
	"gorm.io/gorm"
)

// MessageAttachmentStore 实现MessageAttachmentStore接口的MySQL实现
type MessageAttachmentStore struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewMessageAttachmentStore 创建MySQL消息附件关联存储库实例
func NewMessageAttachmentStore(db *gorm.DB) interfaces.MessageAttachmentStore {
	return &MessageAttachmentStore{db: db}
}

// SetLogger 设置日志记录器
func (r *MessageAttachmentStore) SetLogger(logger *logger.Logger) {
	r.logger = logger
}

// Create 创建消息与附件的关联
func (r *MessageAttachmentStore) Create(messageAttachment *models.MessageAttachment) error {
	err := r.db.Create(messageAttachment).Error
	if err == nil && r.logger != nil {
		r.logger.Info("创建消息 %s 与附件 %s 的关联成功",
			messageAttachment.MessageID, messageAttachment.AttachmentID)
	}
	return err
}

// Delete 根据ID删除消息与附件的关联
func (r *MessageAttachmentStore) Delete(id uint64) error {
	// 根据ID删除记录
	err := r.db.Where("id = ?", id).Delete(&models.MessageAttachment{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("删除ID为 %d 的消息-附件关联成功", id)
	}
	return err
}

// DeleteByMessageAndAttachment 根据消息ID和附件ID删除关联
func (r *MessageAttachmentStore) DeleteByMessageAndAttachment(messageID, attachmentID string) error {
	err := r.db.Where("message_id = ? AND attachment_id = ?", messageID, attachmentID).
		Delete(&models.MessageAttachment{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("删除消息 %s 与附件 %s 的关联成功", messageID, attachmentID)
	}
	return err
}

// DeleteByMessage 删除消息的所有附件关联
func (r *MessageAttachmentStore) DeleteByMessage(messageID string) error {
	err := r.db.Where("message_id = ?", messageID).Delete(&models.MessageAttachment{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("删除消息 %s 的所有附件关联成功", messageID)
	}
	return err
}

// DeleteByAttachment 删除附件的所有消息关联
func (r *MessageAttachmentStore) DeleteByAttachment(attachmentID string) error {
	err := r.db.Where("attachment_id = ?", attachmentID).Delete(&models.MessageAttachment{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("删除附件 %s 的所有消息关联成功", attachmentID)
	}
	return err
}

// ListByMessage 根据消息ID获取消息附件关联列表
func (r *MessageAttachmentStore) ListByMessage(messageID string) ([]*models.MessageAttachment, error) {
	var messageAttachments []*models.MessageAttachment
	err := r.db.Where("message_id = ?", messageID).Find(&messageAttachments).Error
	if err == nil && r.logger != nil {
		r.logger.Debug("查询到消息 %s 的 %d 个附件关联", messageID, len(messageAttachments))
	}
	return messageAttachments, err
}

// ListByAttachment 根据附件ID获取消息附件关联列表
func (r *MessageAttachmentStore) ListByAttachment(attachmentID string) ([]*models.MessageAttachment, error) {
	var messageAttachments []*models.MessageAttachment
	err := r.db.Where("attachment_id = ?", attachmentID).Find(&messageAttachments).Error
	if err == nil && r.logger != nil {
		r.logger.Debug("查询到附件 %s 的 %d 个消息关联", attachmentID, len(messageAttachments))
	}
	return messageAttachments, err
}
