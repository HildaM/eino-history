package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/interfaces"
)

const (
	AttachmentPrefix = "attachment:"
)

// AttachmentStore Redis implementation
type AttachmentStore struct {
	client *redis.Client
	debug  bool
}

// NewAttachmentStore creates a new Redis attachment Store instance
func NewAttachmentStore(client *redis.Client, debug bool) interfaces.AttachmentStore {
	return &AttachmentStore{
		client: client,
		debug:  debug,
	}
}

// Create creates an attachment
func (r *AttachmentStore) Create(attachment *models.Attachment) error {
	ctx := context.Background()

	if len(attachment.AttachID) == 0 {
		attachment.AttachID = uuid.NewString()
	}

	// 转换为JSON
	data, err := json.Marshal(attachment)
	if err != nil {
		if r.debug {
			log.Printf("Redis错误: 附件序列化失败: %v", err)
		}
		return err
	}

	// 存储附件
	key := AttachmentPrefix + attachment.AttachID
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		if r.debug {
			log.Printf("Redis错误: 保存附件失败: %v", err)
		}
		return err
	}

	return nil
}

// Update updates an attachment
func (r *AttachmentStore) Update(attachment *models.Attachment) error {
	ctx := context.Background()

	// 转换为JSON
	data, err := json.Marshal(attachment)
	if err != nil {
		return err
	}

	// 更新附件
	key := AttachmentPrefix + attachment.AttachID
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}

	return nil
}

// Delete deletes an attachment
func (r *AttachmentStore) Delete(attachID string) error {
	ctx := context.Background()

	// 删除附件
	key := AttachmentPrefix + attachID
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

// GetByID gets an attachment by ID
func (r *AttachmentStore) GetByID(attachID string) (*models.Attachment, error) {
	ctx := context.Background()

	key := AttachmentPrefix + attachID
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, err
	}

	var attachment models.Attachment
	if err := json.Unmarshal(data, &attachment); err != nil {
		return nil, err
	}

	return &attachment, nil
}

// ListByMessage gets attachments by message ID
func (r *AttachmentStore) ListByMessage(messageID string) ([]*models.Attachment, error) {
	ctx := context.Background()

	// 获取消息的所有附件关联
	messageKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, messageID)
	ids, err := r.client.SMembers(ctx, messageKey).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*models.Attachment{}, nil
	}

	// 获取每个附件
	var attachments []*models.Attachment
	for _, idStr := range ids {
		// 先获取消息附件关联
		maKey := fmt.Sprintf("%s%s", MessageAttachmentPrefix, idStr)
		data, err := r.client.Get(ctx, maKey).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		var ma models.MessageAttachment
		if err := json.Unmarshal(data, &ma); err != nil {
			return nil, err
		}

		// 获取附件信息
		attachment, err := r.GetByID(ma.AttachmentID)
		if err != nil {
			if err.Error() == "attachment not found" {
				continue
			}
			return nil, err
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}
