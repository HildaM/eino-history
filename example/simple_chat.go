package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hildam/eino-history/eino"
	"github.com/hildam/eino-history/model"
)

// SimpleChat 演示简单的聊天功能
func SimpleChat(ctx context.Context, historyStore *eino.History) {
	log.Println("\n\n===== 简单聊天示例 =====")

	// 创建语言模型客户端
	cm := createLLMChatModel(ctx)

	// 创建对话
	convID := uuid.NewString()
	conv := &models.Conversation{
		ConvID:    convID,
		Title:     "简单聊天示例",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	// 保存对话
	if err := historyStore.CreateConversation(conv); err != nil {
		log.Printf("创建对话失败: %v", err)
		return
	}

	log.Printf("\n开始对话 (ID: %s)\n", convID)

	// 示例问题列表
	questions := []string{
		"你好，请介绍一下你自己",
		"你能帮我解决什么问题？",
		"谢谢你的帮助",
	}

	// 处理对话
	for i, question := range questions {
		log.Printf("\n\n=== 处理消息 #%d ===", i+1)
		log.Printf("问题: %s\n", question)

		// 创建消息
		messages, err := createMessagesWithStore(ctx, convID, question, historyStore)
		if err != nil {
			log.Printf("创建消息失败: %v", err)
			return
		}

		// 记录发送给模型的消息列表
		log.Printf("\n=== 发送给模型的消息列表 ===")
		for j, msg := range messages {
			// 只截断历史消息（用户消息和系统消息），不截断当前问题
			content := msg.Content
			if j < len(messages)-1 && len(content) > 100 {
				content = content[:100] + fmt.Sprintf("... (已截断，共 %d 字符)", len(content))
			}
			log.Printf("[%d] %s: %s", j+1, msg.Role, content)
		}
		log.Printf("=== 消息列表结束 ===\n")

		// 生成回复
		result := generateResponse(ctx, cm, messages)

		// 保存到历史记录中
		err = historyStore.SaveMessage(result, convID)
		if err != nil {
			log.Printf("保存助手消息失败: %v", err)
			return
		}

		// 记录完整回复（不截断）
		log.Printf("\n回复: %s\n", result.Content)
		log.Printf("=== 消息 #%d 处理完成 ===\n", i+1)
	}

	log.Println("\n\n===== 简单聊天示例结束 =====\n\n\n\n")
}
