package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/interfaces"
)

const (
	ConversationKeyPrefix = "conversation:"
)

// ConversationStore 实现ConversationStore接口的Redis实现
type ConversationStore struct {
	client *redis.Client
	debug  bool
}

// NewConversationStore 创建Redis会话存储库实例
func NewConversationStore(client *redis.Client, debug bool) interfaces.ConversationStore {
	return &ConversationStore{
		client: client,
		debug:  debug,
	}
}

// Create 创建会话
func (r *ConversationStore) Create(conv *models.Conversation) error {
	ctx := context.Background()

	if conv.CreatedAt == 0 {
		conv.CreatedAt = time.Now().Unix()
	}

	if conv.UpdatedAt == 0 {
		conv.UpdatedAt = time.Now().Unix()
	}

	// 转换会话为JSON
	data, err := json.Marshal(conv)
	if err != nil {
		return err
	}

	// 存储会话
	key := ConversationKeyPrefix + conv.ConvID
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}

	// 添加到有序集合供列表查询
	if err := r.client.ZAdd(ctx, "conversations", &redis.Z{
		Score:  float64(conv.UpdatedAt),
		Member: conv.ConvID,
	}).Err(); err != nil {
		return err
	}

	if r.debug {
		log.Printf("Redis: 创建会话 %s 成功", conv.ConvID)
	}
	return nil
}

// Update 更新会话
func (r *ConversationStore) Update(conv *models.Conversation) error {
	ctx := context.Background()

	conv.UpdatedAt = time.Now().Unix()

	// 转换会话为JSON
	data, err := json.Marshal(conv)
	if err != nil {
		return err
	}

	// 更新会话
	key := ConversationKeyPrefix + conv.ConvID
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}

	// 更新有序集合中的分数
	if err := r.client.ZAdd(ctx, "conversations", &redis.Z{
		Score:  float64(conv.UpdatedAt),
		Member: conv.ConvID,
	}).Err(); err != nil {
		return err
	}

	return nil
}

// Delete 删除会话
func (r *ConversationStore) Delete(convID string) error {
	ctx := context.Background()

	// 删除会话
	key := ConversationKeyPrefix + convID
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	// 从有序集合中移除
	if err := r.client.ZRem(ctx, "conversations", convID).Err(); err != nil {
		return err
	}

	// 删除会话中的所有消息
	messagesKey := ConversationMessagesPrefix + convID
	messageIDs, err := r.client.ZRange(ctx, messagesKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, msgID := range messageIDs {
		if err := r.client.Del(ctx, MessageKeyPrefix+msgID).Err(); err != nil {
			return err
		}
	}

	// 删除会话消息列表
	if err := r.client.Del(ctx, messagesKey).Err(); err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取会话
func (r *ConversationStore) GetByID(convID string) (*models.Conversation, error) {
	ctx := context.Background()

	key := ConversationKeyPrefix + convID
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, err
	}

	var conv models.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, err
	}

	return &conv, nil
}

// FirstOrCreat 根据ID查找会话，如果不存在则创建
func (r *ConversationStore) FirstOrCreat(convID string) (*models.Conversation, error) {
	// 尝试获取已存在的会话
	conv, err := r.GetByID(convID)
	if err == nil {
		if r.debug {
			log.Printf("Redis: 找到现有会话 %s", convID)
		}
		return conv, nil
	}

	if r.debug {
		log.Printf("Redis: 会话 %s 不存在，创建新会话", convID)
	}

	// 创建新会话
	now := time.Now().Unix()
	newConv := &models.Conversation{
		ConvID:    convID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := r.Create(newConv); err != nil {
		return nil, err
	}

	return newConv, nil
}

// List 获取会话列表
func (r *ConversationStore) List(offset, limit int) ([]*models.Conversation, error) {
	ctx := context.Background()

	// 获取会话ID列表，按UpdatedAt降序排序
	convIDs, err := r.client.ZRevRange(ctx, "conversations", int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, err
	}

	if r.debug {
		log.Printf("Redis: 查询到 %d 个会话", len(convIDs))
	}

	if len(convIDs) == 0 {
		return []*models.Conversation{}, nil
	}

	// 获取每个会话
	var convs []*models.Conversation
	for _, convID := range convIDs {
		conv, err := r.GetByID(convID)
		if err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}

	return convs, nil
}

// Archive 归档会话
func (r *ConversationStore) Archive(convID string) error {
	conv, err := r.GetByID(convID)
	if err != nil {
		return err
	}

	conv.IsArchived = true
	return r.Update(conv)
}

// Unarchive 取消归档会话
func (r *ConversationStore) Unarchive(convID string) error {
	conv, err := r.GetByID(convID)
	if err != nil {
		return err
	}

	conv.IsArchived = false
	return r.Update(conv)
}

// Pin 置顶会话
func (r *ConversationStore) Pin(convID string) error {
	conv, err := r.GetByID(convID)
	if err != nil {
		return err
	}

	conv.IsPinned = true
	return r.Update(conv)
}

// Unpin 取消置顶会话
func (r *ConversationStore) Unpin(convID string) error {
	conv, err := r.GetByID(convID)
	if err != nil {
		return err
	}

	conv.IsPinned = false
	return r.Update(conv)
}
