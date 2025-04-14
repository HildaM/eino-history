package mysql

import (
	"github.com/google/uuid"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
	"gorm.io/gorm"
)

// AttachmentStore 实现AttachmentStore接口的MySQL实现
type AttachmentStore struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewAttachmentStore 创建MySQL附件存储库实例
func NewAttachmentStore(db *gorm.DB) interfaces.AttachmentStore {
	return &AttachmentStore{db: db}
}

// SetLogger 设置日志记录器
func (r *AttachmentStore) SetLogger(logger *logger.Logger) {
	r.logger = logger
}

// Create 创建附件
func (r *AttachmentStore) Create(attachment *models.Attachment) error {
	if len(attachment.AttachID) == 0 {
		attachment.AttachID = uuid.NewString()
	}

	err := r.db.Create(attachment).Error
	if err == nil && r.logger != nil {
		r.logger.Info("附件 %s 创建成功", attachment.AttachID)
	}
	return err
}

// Update 更新附件
func (r *AttachmentStore) Update(attachment *models.Attachment) error {
	err := r.db.Save(attachment).Error
	if err == nil && r.logger != nil {
		r.logger.Info("附件 %s 更新成功", attachment.AttachID)
	}
	return err
}

// Delete 删除附件
func (r *AttachmentStore) Delete(attachID string) error {
	err := r.db.Where("attach_id = ?", attachID).Delete(&models.Attachment{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("附件 %s 删除成功", attachID)
	}
	return err
}

// GetByID 根据ID获取附件
func (r *AttachmentStore) GetByID(attachID string) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.Where("attach_id = ?", attachID).First(&attachment).Error
	if err != nil && r.logger != nil {
		r.logger.Error("获取附件 %s 失败: %v", attachID, err)
		return nil, err
	}
	return &attachment, nil
}

// ListByMessage 获取消息的附件列表
func (r *AttachmentStore) ListByMessage(messageID string) ([]*models.Attachment, error) {
	// 使用关联表查询
	var messageAttachments []*models.MessageAttachment
	err := r.db.Where("message_id = ?", messageID).Find(&messageAttachments).Error
	if err != nil && r.logger != nil {
		r.logger.Error("查询消息 %s 的附件关联失败: %v", messageID, err)
		return nil, err
	}

	// 如果没有附件，直接返回空列表
	if len(messageAttachments) == 0 {
		if r.logger != nil {
			r.logger.Debug("消息 %s 没有附件", messageID)
		}
		return []*models.Attachment{}, nil
	}

	// 提取附件ID列表
	var attachmentIDs []string
	for _, ma := range messageAttachments {
		attachmentIDs = append(attachmentIDs, ma.AttachmentID)
	}

	// 查询所有附件
	var attachments []*models.Attachment
	err = r.db.Where("attach_id IN ?", attachmentIDs).Find(&attachments).Error

	if err == nil && r.logger != nil {
		r.logger.Info("查询到消息 %s 的 %d 个附件", messageID, len(attachments))
	}

	return attachments, err
}
