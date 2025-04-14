package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log" // 暂时保留用于兼容
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/interfaces"
)

// Redis key patterns
const (
	MessageKeyPrefix           = "message:"
	ConversationMessagesPrefix = "conversation:messages:"
)

// MessageStore 实现MessageStore接口的Redis实现
type MessageStore struct {
	client *redis.Client
	debug  bool
	logger *logger.Logger
}

// NewMessageStore 创建Redis消息存储库实例
func NewMessageStore(client *redis.Client, debug bool) interfaces.MessageStore {
	return &MessageStore{
		client: client,
		debug:  debug,
	}
}

// SetLogger 设置日志记录器
func (r *MessageStore) SetLogger(logger *logger.Logger) {
	r.logger = logger
}

// Create 创建消息
func (r *MessageStore) Create(msg *models.Message) error {
	ctx := context.Background()

	if len(msg.MsgID) == 0 {
		msg.MsgID = uuid.NewString()
	}

	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().Unix()
	}

	// 转换消息为JSON
	data, err := json.Marshal(msg)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("消息序列化失败: %v", err)
		} else if r.debug {
			// 兼容旧的日志记录方式，将来可以移除
			r.logError("消息序列化失败: %v", err)
		}
		return err
	}

	// 存储消息
	key := MessageKeyPrefix + msg.MsgID
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("保存消息失败: %v", err)
		} else if r.debug {
			r.logError("保存消息失败: %v", err)
		}
		return err
	}

	// 将消息添加到会话列表
	convKey := ConversationMessagesPrefix + msg.ConversationID
	if err := r.client.ZAdd(ctx, convKey, &redis.Z{
		Score:  float64(msg.OrderSeq),
		Member: msg.MsgID,
	}).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("添加消息到会话列表失败: %v", err)
		} else if r.debug {
			r.logError("添加消息到会话列表失败: %v", err)
		}
		return err
	}

	if r.logger != nil {
		r.logger.Info("消息 %s 创建成功", msg.MsgID)
	}

	return nil
}

// 兼容旧版本的日志输出，将来可以移除
func (r *MessageStore) logError(format string, args ...interface{}) {
	log.Printf("Redis错误: "+format, args...)
}

// Update 更新消息
func (r *MessageStore) Update(msg *models.Message) error {
	ctx := context.Background()

	// 转换消息为JSON
	data, err := json.Marshal(msg)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("消息序列化失败: %v", err)
		} else if r.debug {
			r.logError("消息序列化失败: %v", err)
		}
		return err
	}

	// 更新消息
	key := MessageKeyPrefix + msg.MsgID
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("更新消息失败: %v", err)
		} else if r.debug {
			r.logError("更新消息失败: %v", err)
		}
		return err
	}

	// 更新会话列表中的排序分数
	convKey := ConversationMessagesPrefix + msg.ConversationID
	if err := r.client.ZAdd(ctx, convKey, &redis.Z{
		Score:  float64(msg.OrderSeq),
		Member: msg.MsgID,
	}).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("更新消息排序分数失败: %v", err)
		} else if r.debug {
			r.logError("更新消息排序分数失败: %v", err)
		}
		return err
	}

	if r.logger != nil {
		r.logger.Info("消息 %s 更新成功", msg.MsgID)
	}

	return nil
}

// Delete 删除消息
func (r *MessageStore) Delete(msgID string) error {
	ctx := context.Background()

	// 先获取消息以便知道它属于哪个会话
	msg, err := r.GetByID(msgID)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("获取消息失败: %v", err)
		} else if r.debug {
			r.logError("获取消息失败: %v", err)
		}
		return err
	}

	// 删除消息
	key := MessageKeyPrefix + msgID
	if err := r.client.Del(ctx, key).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("删除消息失败: %v", err)
		} else if r.debug {
			r.logError("删除消息失败: %v", err)
		}
		return err
	}

	// 从会话列表中移除
	convKey := ConversationMessagesPrefix + msg.ConversationID
	if err := r.client.ZRem(ctx, convKey, msgID).Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("从会话中移除消息失败: %v", err)
		} else if r.debug {
			r.logError("从会话中移除消息失败: %v", err)
		}
		return err
	}

	if r.logger != nil {
		r.logger.Info("消息 %s 删除成功", msgID)
	}

	return nil
}

// GetByID 根据ID获取消息
func (r *MessageStore) GetByID(msgID string) (*models.Message, error) {
	ctx := context.Background()

	key := MessageKeyPrefix + msgID
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			if r.logger != nil {
				r.logger.Error("消息不存在: %s", msgID)
			} else if r.debug {
				r.logError("消息不存在: %s", msgID)
			}
			return nil, fmt.Errorf("message not found")
		}
		if r.logger != nil {
			r.logger.Error("获取消息失败: %v", err)
		} else if r.debug {
			r.logError("获取消息失败: %v", err)
		}
		return nil, err
	}

	var msg models.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		if r.logger != nil {
			r.logger.Error("消息反序列化失败: %v", err)
		} else if r.debug {
			r.logError("消息反序列化失败: %v", err)
		}
		return nil, err
	}

	return &msg, nil
}

// ListByConversation 获取对话的消息列表
func (r *MessageStore) ListByConversation(conversationID string, offset, limit int) ([]*models.Message, error) {
	ctx := context.Background()

	convKey := ConversationMessagesPrefix + conversationID

	// 获取消息ID列表，按OrderSeq排序
	msgIDs, err := r.client.ZRange(ctx, convKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("获取会话消息ID列表失败: %v", err)
		} else if r.debug {
			r.logError("获取会话消息ID列表失败: %v", err)
		}
		return nil, err
	}

	if r.logger != nil {
		r.logger.Info("会话 %s 查询到 %d 条消息", conversationID, len(msgIDs))
	} else if r.debug {
		log.Printf("Redis: 会话 %s 查询到 %d 条消息", conversationID, len(msgIDs))
	}

	if len(msgIDs) == 0 {
		return []*models.Message{}, nil
	}

	// 获取每条消息
	var msgs []*models.Message
	for _, msgID := range msgIDs {
		msg, err := r.GetByID(msgID)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("获取消息 %s 详情失败: %v", msgID, err)
			} else if r.debug {
				r.logError("获取消息 %s 详情失败: %v", msgID, err)
			}
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	// 打印消息内容摘要，方便调试
	if r.logger != nil && r.logger.Level >= logger.DebugLevel {
		for i, msg := range msgs {
			if len(msg.Content) > 50 {
				r.logger.Debug("消息[%d] 角色=%s, 内容=%s...", i, msg.Role, msg.Content[:50])
			} else {
				r.logger.Debug("消息[%d] 角色=%s, 内容=%s", i, msg.Role, msg.Content)
			}
		}
	} else if r.debug {
		for i, msg := range msgs {
			if len(msg.Content) > 50 {
				log.Printf("Redis: 消息[%d] 角色=%s, 内容=%s...", i, msg.Role, msg.Content[:50])
			} else {
				log.Printf("Redis: 消息[%d] 角色=%s, 内容=%s", i, msg.Role, msg.Content)
			}
		}
	}

	return msgs, nil
}

// UpdateStatus 更新消息状态
func (r *MessageStore) UpdateStatus(msgID string, status string) error {
	msg, err := r.GetByID(msgID)
	if err != nil {
		return err
	}

	msg.Status = status
	return r.Update(msg)
}

// UpdateTokenCount 更新消息token数量
func (r *MessageStore) UpdateTokenCount(msgID string, tokenCount int) error {
	msg, err := r.GetByID(msgID)
	if err != nil {
		return err
	}

	msg.TokenCount = tokenCount
	return r.Update(msg)
}

// SetContextEdge 设置消息为上下文边界
func (r *MessageStore) SetContextEdge(msgID string, isContextEdge bool) error {
	msg, err := r.GetByID(msgID)
	if err != nil {
		return err
	}

	msg.IsContextEdge = isContextEdge
	return r.Update(msg)
}

// SetVariant 设置消息为变体
func (r *MessageStore) SetVariant(msgID string, isVariant bool) error {
	msg, err := r.GetByID(msgID)
	if err != nil {
		return err
	}

	msg.IsVariant = isVariant
	return r.Update(msg)
}
