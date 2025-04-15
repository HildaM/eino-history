package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/google/uuid"
	"github.com/hildam/eino-history/conf"
	"github.com/hildam/eino-history/eino"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/provider"
)

// 初始化 MySQL 历史记录
var ehMySQL = eino.NewDefaultEinoHistory("root:123456@tcp(127.0.0.1:3306)/chat_history?charset=utf8mb4&parseTime=True&loc=Local", "info")

// 初始化 Redis 历史记录 (Redis URL 格式: redis://user:password@localhost:6379/0)
var ehRedis = eino.NewEinoHistoryWithProvider("redis://localhost:6379/0", provider.TypeRedis, true, "info")

func main() {
	// 确保程序结束时关闭数据库连接
	defer func() {
		if err := ehMySQL.Close(); err != nil {
			log.Printf("关闭 MySQL 连接失败: %v", err)
		}
		if err := ehRedis.Close(); err != nil {
			log.Printf("关闭 Redis 连接失败: %v", err)
		}
	}()

	ctx := context.Background()
	if err := conf.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 解析命令行参数
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// 根据参数执行不同的示例
	switch os.Args[1] {
	case "simple":
		// 简单聊天示例
		log.Println("执行简单聊天示例 (MySQL)...")
		SimpleChat(ctx, ehMySQL)

	case "redis-simple":
		// 使用Redis的简单聊天示例
		log.Println("执行简单聊天示例 (Redis)...")
		SimpleChat(ctx, ehRedis)

	case "conversation":
		// 对话管理示例
		log.Println("执行对话管理示例 (MySQL)...")
		ConversationManagement(ctx, ehMySQL)

	case "redis-conversation":
		// 使用Redis的对话管理示例
		log.Println("执行对话管理示例 (Redis)...")
		ConversationManagement(ctx, ehRedis)

	case "attachment":
		// 附件功能示例
		log.Println("执行附件功能示例...")
		AttachmentExample()

	case "all":
		// 运行所有示例
		log.Println("执行所有示例...")

		log.Println("\n1. 简单聊天示例 (MySQL)")
		SimpleChat(ctx, ehMySQL)

		log.Println("\n2. 简单聊天示例 (Redis)")
		SimpleChat(ctx, ehRedis)

		log.Println("\n3. 对话管理示例 (MySQL)")
		ConversationManagement(ctx, ehMySQL)

		log.Println("\n4. 对话管理示例 (Redis)")
		ConversationManagement(ctx, ehRedis)

		log.Println("\n5. 附件功能示例")
		AttachmentExample()

	default:
		fmt.Printf("未知的示例: %s\n", os.Args[1])
		printUsage()
	}
}

// printUsage 打印使用帮助
func printUsage() {
	fmt.Println("使用方法: go run *.go <示例名称>")
	fmt.Println("可用的示例:")
	fmt.Println("  simple            - 简单聊天示例 (MySQL)")
	fmt.Println("  redis-simple      - 简单聊天示例 (Redis)")
	fmt.Println("  conversation      - 对话管理示例 (MySQL)")
	fmt.Println("  redis-conversation - 对话管理示例 (Redis)")
	fmt.Println("  attachment        - 附件功能示例")
	fmt.Println("  all               - 运行所有示例")
}

// processConversationWithTopic 处理带主题的对话
func processConversationWithTopic(ctx context.Context, topics, messList []string, historyStore *eino.History) {
	cm := createLLMChatModel(ctx)

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

		log.Printf("开始对话: %s (ID: %s)", topic, convID)

		// 处理对话
		processConversation(ctx, cm, convID, messList, historyStore)

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
			}
		} else if i == 1 {
			// 置顶第二个对话
			if err := historyStore.PinConversation(convID); err != nil {
				log.Printf("置顶对话失败: %v", err)
			}
		}

		// 获取并显示对话列表
		convs, err := historyStore.ListConversations(0, 10)
		if err != nil {
			log.Printf("获取对话列表失败: %v", err)
		} else {
			log.Printf("当前对话列表:")
			for _, c := range convs {
				log.Printf("- %s (ID: %s, 归档: %v, 置顶: %v)", c.Title, c.ConvID, c.IsArchived, c.IsPinned)
			}
		}
	}
}

// processConversation 处理一个完整的对话，使用指定的历史记录存储
func processConversation(ctx context.Context, cm model.ChatModel, convID string, messList []string, historyStore *eino.History) {
	for _, s := range messList {
		// 创建消息，从指定的历史记录中获取
		messages, err := createMessagesWithStore(ctx, convID, s, historyStore)
		if err != nil {
			log.Fatalf("create messages failed: %v", err)
			return
		}

		// 生成回复
		result := generateResponse(ctx, cm, messages)

		// 保存到指定的历史记录中
		err = historyStore.SaveMessage(result, convID)
		if err != nil {
			log.Fatalf("save assistant message err: %v", err)
			return
		}

		log.Printf("result: %+v\n\n", result)
	}
}

// attachmentExample 演示如何使用附件功能
func attachmentExample() {
	// 创建Redis连接配置
	config := &provider.Config{
		DSN:      "redis://localhost:6379/0",
		Type:     provider.TypeRedis,
		Debug:    true,
		LogLevel: "debug",
	}

	// 创建数据库提供者实例
	dbProvider, err := provider.CreateProvider(config)
	if err != nil {
		log.Fatalf("创建数据库提供者失败: %v", err)
	}
	defer dbProvider.Close() // 确保最后关闭连接

	// 获取各种仓库
	messageRepo := dbProvider.GetMessageStore()
	attachmentRepo := dbProvider.GetAttachmentStore()
	messageAttachmentRepo := dbProvider.GetMessageAttachmentStore()

	// 创建一个测试会话ID
	conversationID := uuid.NewString()
	log.Printf("创建会话 ID: %s", conversationID)

	// 创建一条消息
	message := &models.Message{
		MsgID:          uuid.NewString(),
		ConversationID: conversationID,
		Role:           "user",
		Content:        "这是一条带有附件的消息",
		OrderSeq:       1,
		CreatedAt:      time.Now().Unix(),
	}

	if err := messageRepo.Create(message); err != nil {
		log.Fatalf("创建消息失败: %v", err)
	}
	log.Printf("创建消息成功 - ID: %s, 内容: %s", message.MsgID, message.Content)

	// 创建两个不同类型的附件
	attachments := []*models.Attachment{
		{
			AttachID:       uuid.NewString(),
			AttachmentType: "image",
			FileName:       "example.jpg",
			FileSize:       1024 * 100, // 100KB
			StorageType:    "path",
			StoragePath:    "/images/example.jpg",
			MimeType:       "image/jpeg",
			CreatedAt:      time.Now().Unix(),
		},
		{
			AttachID:       uuid.NewString(),
			AttachmentType: "file",
			FileName:       "document.pdf",
			FileSize:       1024 * 1024 * 2, // 2MB
			StorageType:    "path",
			StoragePath:    "/documents/example.pdf",
			MimeType:       "application/pdf",
			CreatedAt:      time.Now().Unix(),
		},
	}

	// 保存附件并将其关联到消息
	for _, attachment := range attachments {
		// 保存附件
		if err := attachmentRepo.Create(attachment); err != nil {
			log.Fatalf("创建附件失败: %v", err)
		}
		log.Printf("创建附件成功 - ID: %s, 名称: %s, 类型: %s",
			attachment.AttachID, attachment.FileName, attachment.AttachmentType)

		// 创建消息与附件的关联
		messageAttachment := &models.MessageAttachment{
			MessageID:    message.MsgID,
			AttachmentID: attachment.AttachID,
		}

		if err := messageAttachmentRepo.Create(messageAttachment); err != nil {
			log.Fatalf("关联附件到消息失败: %v", err)
		}
		log.Printf("附件 %s 已关联到消息 %s", attachment.AttachID, message.MsgID)
	}

	// 获取消息的所有附件关联
	messageAttachments, err := messageAttachmentRepo.ListByMessage(message.MsgID)
	if err != nil {
		log.Fatalf("获取消息附件关联失败: %v", err)
	}
	log.Printf("消息 %s 有 %d 个附件关联", message.MsgID, len(messageAttachments))

	// 获取消息的所有附件详情
	attachmentList, err := attachmentRepo.ListByMessage(message.MsgID)
	if err != nil {
		log.Fatalf("获取消息附件详情失败: %v", err)
	}

	log.Printf("附件详情列表:")
	for i, attachment := range attachmentList {
		log.Printf("附件 #%d:", i+1)
		log.Printf("  ID: %s", attachment.AttachID)
		log.Printf("  文件名: %s", attachment.FileName)
		log.Printf("  类型: %s", attachment.AttachmentType)
		log.Printf("  大小: %d 字节", attachment.FileSize)
		log.Printf("  MIME类型: %s", attachment.MimeType)
		log.Printf("  存储路径: %s", attachment.StoragePath)
	}

	// 测试获取单个附件
	if len(attachmentList) > 0 {
		firstAttachment, err := attachmentRepo.GetByID(attachmentList[0].AttachID)
		if err != nil {
			log.Fatalf("获取单个附件失败: %v", err)
		}
		log.Printf("获取单个附件成功 - ID: %s, 名称: %s",
			firstAttachment.AttachID, firstAttachment.FileName)
	}

	log.Println("附件示例执行完成")
}
