package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/hildam/eino-history/conf"
	"github.com/hildam/eino-history/eino"
)

// createLLMChatModel 创建语言模型客户端
func createLLMChatModel(ctx context.Context) model.ChatModel {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   conf.GetCfg().DeekSeek.ModelID,
		APIKey:  conf.GetCfg().DeekSeek.APIKey,
		BaseURL: conf.GetCfg().DeekSeek.BaseURL,
	})
	if err != nil {
		log.Fatalf("create openai chat model failed, err=%v", err)
	}
	return chatModel
}

// generateResponse 使用聊天模型生成回复
func generateResponse(ctx context.Context, cm model.ChatModel, messages []schema.Message) *schema.Message {
	// 将 []schema.Message 转换为 []*schema.Message 以便传入 Generate 方法
	var messagePtrs []*schema.Message
	for i := range messages {
		messagePtrs = append(messagePtrs, &messages[i])
	}

	resp, err := cm.Generate(ctx, messagePtrs)
	if err != nil {
		log.Printf("chat failed: %v", err)
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "Sorry, I encountered an error while processing your request.",
		}
	}

	return resp
}

// createMessagesWithStore 使用指定的历史记录存储创建消息列表
func createMessagesWithStore(ctx context.Context, convID, userInput string, historyStore *eino.History) ([]schema.Message, error) {
	// 从指定的历史记录中获取之前的消息
	historyMessages, err := historyStore.GetHistory(convID, 100)
	if err != nil {
		return nil, err
	}

	// 添加新的用户消息
	userMessage := schema.Message{
		Role:    schema.User,
		Content: userInput,
	}

	// 保存用户消息到指定的历史记录中
	err = historyStore.SaveMessage(&userMessage, convID)
	if err != nil {
		return nil, err
	}

	// 合并历史消息和新消息
	var allMessages []schema.Message
	for _, msg := range historyMessages {
		allMessages = append(allMessages, *msg)
	}
	allMessages = append(allMessages, userMessage)

	return allMessages, nil
}
