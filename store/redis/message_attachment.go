package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/interfaces"
)

const (
	MessageAttachmentPrefix = "message_attachment:"
	MessageAttachmentsKey   = "message_attachments"
)

// MessageAttachmentStore Redis implementation
type MessageAttachmentStore struct {
	client *redis.Client
	debug  bool
}

// NewMessageAttachmentStore creates a new Redis message attachment Store instance
func NewMessageAttachmentStore(client *redis.Client, debug bool) interfaces.MessageAttachmentStore {
	return &MessageAttachmentStore{
		client: client,
		debug:  debug,
	}
}

// Create creates a message attachment association
func (r *MessageAttachmentStore) Create(messageAttachment *models.MessageAttachment) error {
	ctx := context.Background()

	// 转换为JSON
	data, err := json.Marshal(messageAttachment)
	if err != nil {
		if r.debug {
			log.Printf("Redis错误: 消息附件序列化失败: %v", err)
		}
		return err
	}

	// 存储消息附件关联
	key := fmt.Sprintf("%s%d", MessageAttachmentPrefix, messageAttachment.ID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		if r.debug {
			log.Printf("Redis错误: 保存消息附件关联失败: %v", err)
		}
		return err
	}

	// 添加到消息的附件集合
	messageKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, messageAttachment.MessageID)
	if err := r.client.SAdd(ctx, messageKey, messageAttachment.ID).Err(); err != nil {
		return err
	}

	// 添加到附件的消息集合
	attachmentKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, messageAttachment.AttachmentID)
	if err := r.client.SAdd(ctx, attachmentKey, messageAttachment.ID).Err(); err != nil {
		return err
	}

	return nil
}

// Delete deletes a message attachment association
func (r *MessageAttachmentStore) Delete(id uint64) error {
	ctx := context.Background()

	// 先获取消息附件关联信息
	key := fmt.Sprintf("%s%d", MessageAttachmentPrefix, id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("message attachment not found")
		}
		return err
	}

	var messageAttachment models.MessageAttachment
	if err := json.Unmarshal(data, &messageAttachment); err != nil {
		return err
	}

	// 从消息的附件集合中移除
	messageKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, messageAttachment.MessageID)
	if err := r.client.SRem(ctx, messageKey, id).Err(); err != nil {
		return err
	}

	// 从附件的消息集合中移除
	attachmentKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, messageAttachment.AttachmentID)
	if err := r.client.SRem(ctx, attachmentKey, id).Err(); err != nil {
		return err
	}

	// 删除消息附件关联
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

// ListByMessage gets message attachment associations by message ID
func (r *MessageAttachmentStore) ListByMessage(messageID string) ([]*models.MessageAttachment, error) {
	ctx := context.Background()

	// 获取消息的所有附件关联ID
	messageKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, messageID)
	ids, err := r.client.SMembers(ctx, messageKey).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*models.MessageAttachment{}, nil
	}

	// 获取每个消息附件关联
	var messageAttachments []*models.MessageAttachment
	for _, idStr := range ids {
		key := fmt.Sprintf("%s%s", MessageAttachmentPrefix, idStr)
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		var messageAttachment models.MessageAttachment
		if err := json.Unmarshal(data, &messageAttachment); err != nil {
			return nil, err
		}
		messageAttachments = append(messageAttachments, &messageAttachment)
	}

	return messageAttachments, nil
}

// ListByAttachment gets message attachment associations by attachment ID
func (r *MessageAttachmentStore) ListByAttachment(attachmentID string) ([]*models.MessageAttachment, error) {
	ctx := context.Background()

	// 获取附件的所有消息关联ID
	attachmentKey := fmt.Sprintf("%s%s", MessageAttachmentsKey, attachmentID)
	ids, err := r.client.SMembers(ctx, attachmentKey).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*models.MessageAttachment{}, nil
	}

	// 获取每个消息附件关联
	var messageAttachments []*models.MessageAttachment
	for _, idStr := range ids {
		key := fmt.Sprintf("%s%s", MessageAttachmentPrefix, idStr)
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		var messageAttachment models.MessageAttachment
		if err := json.Unmarshal(data, &messageAttachment); err != nil {
			return nil, err
		}
		messageAttachments = append(messageAttachments, &messageAttachment)
	}

	return messageAttachments, nil
}
