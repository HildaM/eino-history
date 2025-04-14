package mysql

import (
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
	"gorm.io/gorm"
)

// ConversationStore 实现ConversationStore接口的MySQL实现
type ConversationStore struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewConversationStore 创建MySQL会话存储库实例
func NewConversationStore(db *gorm.DB) interfaces.ConversationStore {
	return &ConversationStore{db: db}
}

// SetLogger 设置日志记录器
func (r *ConversationStore) SetLogger(logger *logger.Logger) {
	r.logger = logger
}

// Create 创建会话
func (r *ConversationStore) Create(conv *models.Conversation) error {
	err := r.db.Create(conv).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 创建成功", conv.ConvID)
	}
	return err
}

// Update 更新会话
func (r *ConversationStore) Update(conv *models.Conversation) error {
	err := r.db.Save(conv).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 更新成功", conv.ConvID)
	}
	return err
}

// Delete 删除会话
func (r *ConversationStore) Delete(convID string) error {
	err := r.db.Where("conv_id = ?", convID).Delete(&models.Conversation{}).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 删除成功", convID)
	}
	return err
}

// GetByID 根据ID获取会话
func (r *ConversationStore) GetByID(convID string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where("conv_id = ?", convID).First(&conv).Error
	if err != nil && r.logger != nil {
		r.logger.Error("获取会话 %s 失败: %v", convID, err)
	}
	return &conv, nil
}

// FirstOrCreat 根据ID查找会话，如果不存在则创建
func (r *ConversationStore) FirstOrCreat(convID string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where(models.Conversation{ConvID: convID}).FirstOrCreate(&conv).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 查找或创建成功", convID)
	}
	return &conv, err
}

// List 获取会话列表
func (r *ConversationStore) List(offset, limit int) ([]*models.Conversation, error) {
	var convs []*models.Conversation
	err := r.db.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&convs).Error
	if err == nil && r.logger != nil {
		r.logger.Info("查询到 %d 个会话记录", len(convs))
	}
	return convs, err
}

// Archive 归档会话
func (r *ConversationStore) Archive(convID string) error {
	err := r.db.Model(&models.Conversation{}).Where("conv_id = ?", convID).Update("is_archived", true).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 已归档", convID)
	}
	return err
}

// Unarchive 取消归档会话
func (r *ConversationStore) Unarchive(convID string) error {
	err := r.db.Model(&models.Conversation{}).Where("conv_id = ?", convID).Update("is_archived", false).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 已取消归档", convID)
	}
	return err
}

// Pin 置顶会话
func (r *ConversationStore) Pin(convID string) error {
	err := r.db.Model(&models.Conversation{}).Where("conv_id = ?", convID).Update("is_pinned", true).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 已置顶", convID)
	}
	return err
}

// Unpin 取消置顶会话
func (r *ConversationStore) Unpin(convID string) error {
	err := r.db.Model(&models.Conversation{}).Where("conv_id = ?", convID).Update("is_pinned", false).Error
	if err == nil && r.logger != nil {
		r.logger.Info("会话 %s 已取消置顶", convID)
	}
	return err
}
