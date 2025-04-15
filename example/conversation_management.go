package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/hildam/eino-history/eino"
	"github.com/hildam/eino-history/model"
)

// ConversationManagement 演示对话管理功能
func ConversationManagement(ctx context.Context, historyStore *eino.History) {
	log.Println("\n\n===== 对话管理示例 =====")

	// 创建语言模型客户端
	cm := createLLMChatModel(ctx)

	// 创建对话主题
	topics := []string{
		"数学学习",
		"编程学习",
		"对话回顾",
	}

	for i, topic := range topics {
		convID := uuid.NewString()

		// 创建对话
		conv := &models.Conversation{
			ConvID:    convID,
			Title:     topic,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}

		// 设置对话元数据
		settings := map[string]interface{}{
			"topic": topic,
			"index": i,
		}
		settingsJSON, _ := json.Marshal(settings)
		conv.Settings = settingsJSON

		// 保存对话
		if err := historyStore.CreateConversation(conv); err != nil {
			log.Printf("创建对话失败: %v", err)
			continue
		}

		log.Printf("\n\n开始对话: %s (ID: %s)\n", topic, convID)

		// 根据主题获取对应的问题列表
		questions := GetQuestionsByTopic(topic)

		// 处理对话
		processConversationMessages(ctx, cm, convID, questions, historyStore)

		// 更新对话状态
		conv.UpdatedAt = time.Now().Unix()
		if err := historyStore.UpdateConversation(conv); err != nil {
			log.Printf("更新对话状态失败: %v", err)
		}

		// 演示归档和置顶功能
		if i == 0 {
			// 归档第一个对话
			if err := historyStore.ArchiveConversation(convID); err != nil {
				log.Printf("归档对话失败: %v", err)
			} else {
				log.Printf("\n对话 '%s' 已归档", topic)
			}
		} else if i == 1 {
			// 置顶第二个对话
			if err := historyStore.PinConversation(convID); err != nil {
				log.Printf("置顶对话失败: %v", err)
			} else {
				log.Printf("\n对话 '%s' 已置顶", topic)
			}
		}
	}

	// 获取并显示对话列表
	convs, err := historyStore.ListConversations(0, 10)
	if err != nil {
		log.Printf("获取对话列表失败: %v", err)
	} else {
		log.Printf("\n\n当前对话列表 (共 %d 个):", len(convs))
		for i, c := range convs {
			log.Printf("#%d - %s (ID: %s, 归档: %v, 置顶: %v)",
				i+1, c.Title, c.ConvID, c.IsArchived, c.IsPinned)
		}
	}

	log.Println("\n\n===== 对话管理示例结束 =====\n")
}

// getRandomHistoricalContext 获取随机历史对话作为上下文
func getRandomHistoricalContext(ctx context.Context, historyStore *eino.History, currentConvID string) ([]*schema.Message, error) {
	// 获取所有对话列表
	convs, err := historyStore.ListConversations(0, 100)
	if err != nil {
		return nil, err
	}

	// 过滤掉当前对话并按时间排序
	var availableConvs []*models.Conversation
	for _, conv := range convs {
		if conv.ConvID != currentConvID {
			availableConvs = append(availableConvs, conv)
		}
	}

	if len(availableConvs) == 0 {
		log.Println("没有可用的历史对话")
		return nil, nil
	}

	// 按更新时间排序（从新到旧）
	sort.Slice(availableConvs, func(i, j int) bool {
		return availableConvs[i].UpdatedAt > availableConvs[j].UpdatedAt
	})

	// 只考虑最近30天内的对话
	var recentConvs []*models.Conversation
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Unix()
	for _, conv := range availableConvs {
		if conv.UpdatedAt >= thirtyDaysAgo {
			recentConvs = append(recentConvs, conv)
		}
	}

	// 如果没有最近的对话，使用所有可用对话
	if len(recentConvs) == 0 {
		log.Println("没有最近30天的对话，使用所有可用对话")
		recentConvs = availableConvs
	}

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 随机选择一个对话
	selectedIndex := rand.Intn(len(recentConvs))
	selectedConv := recentConvs[selectedIndex]

	log.Printf("\n\n=== 历史对话上下文信息 ===")
	log.Printf("选择的对话标题: %s", selectedConv.Title)
	log.Printf("对话ID: %s", selectedConv.ConvID)
	log.Printf("更新时间: %s\n", time.Unix(selectedConv.UpdatedAt, 0).Format("2006-01-02 15:04:05"))

	// 获取该对话的历史消息
	history, err := historyStore.GetHistory(selectedConv.ConvID, 5) // 获取最近5条消息
	if err != nil {
		return nil, err
	}

	// 记录历史消息内容
	log.Printf("\n历史消息内容:")
	for i, msg := range history {
		// 总是截断历史消息
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + fmt.Sprintf("... (已截断，共 %d 字符)", len(content))
		}
		log.Printf("[%d] %s: %s", i+1, msg.Role, content)
	}
	log.Printf("=== 历史对话上下文信息结束 ===\n")

	return history, nil
}

// processConversationMessages 处理一个完整的对话，使用指定的历史记录存储
func processConversationMessages(ctx context.Context, cm model.ChatModel, convID string, messList []string, historyStore *eino.History) {
	// 获取当前对话信息以确定主题
	convs, err := historyStore.ListConversations(0, 100)
	if err != nil {
		log.Printf("获取对话列表失败: %v", err)
		return
	}

	// 查找当前对话
	var currentConv *models.Conversation
	for _, conv := range convs {
		if conv.ConvID == convID {
			currentConv = conv
			break
		}
	}

	if currentConv == nil {
		log.Printf("未找到对话: %s", convID)
		return
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(currentConv.Settings, &settings); err != nil {
		log.Printf("解析对话设置失败: %v", err)
		return
	}

	topic := settings["topic"].(string)
	isReviewTopic := topic == "对话回顾"

	for i, s := range messList {
		log.Printf("\n\n=== 处理消息 #%d ===", i+1)
		log.Printf("问题: %s\n", s)

		// 如果是对话回顾主题，尝试获取历史上下文
		var historicalContext []*schema.Message
		if isReviewTopic {
			historicalContext, err = getRandomHistoricalContext(ctx, historyStore, convID)
			if err != nil {
				log.Printf("获取历史上下文失败: %v", err)
			}
		}

		// 创建消息，从指定的历史记录中获取
		messages, err := createMessagesWithStore(ctx, convID, s, historyStore)
		if err != nil {
			log.Printf("创建消息失败: %v", err)
			return
		}

		// 如果有历史上下文，将其添加到消息列表前面
		if len(historicalContext) > 0 {
			log.Printf("\n=== 最终发送给模型的消息列表 ===")
			// 添加系统消息说明这是历史上下文
			systemMsg := &schema.Message{
				Role:    schema.System,
				Content: "以下是之前的一段对话，请参考这段对话来回答用户的问题：",
			}
			messages = append([]schema.Message{*systemMsg}, messages...)

			// 添加历史消息
			for _, msg := range historicalContext {
				messages = append([]schema.Message{*msg}, messages...)
			}

			// 记录最终发送给模型的消息列表
			for j, msg := range messages {
				// 只截断历史消息，不截断当前问题
				content := msg.Content
				// 计算当前问题在消息列表中的位置
				currentQuestionPosition := len(messages) - 1
				if j != currentQuestionPosition && len(content) > 100 {
					content = content[:100] + fmt.Sprintf("... (已截断，共 %d 字符)", len(content))
				}
				log.Printf("[%d] %s: %s", j+1, msg.Role, content)
			}
			log.Printf("=== 消息列表结束 ===\n")
		}

		// 生成回复
		result := generateResponse(ctx, cm, messages)

		// 保存到指定的历史记录中
		err = historyStore.SaveMessage(result, convID)
		if err != nil {
			log.Printf("保存助手消息失败: %v", err)
			return
		}

		// 记录完整回复（不截断）
		log.Printf("\n回复: %s\n", result.Content)
		log.Printf("=== 消息 #%d 处理完成 ===\n\n\n\n", i+1)
	}
}
