package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/hildam/eino-history/conf"
	"github.com/hildam/eino-history/eino"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/provider"
)

// 初始化 MySQL 历史记录
var ehMySQL = eino.NewDefaultEinoHistory("root:123456@tcp(127.0.0.1:3306)/chat_history?charset=utf8mb4&parseTime=True&loc=Local", "debug")

// 初始化 Redis 历史记录 (Redis URL 格式: redis://user:password@localhost:6379/0)
var ehRedis = eino.NewEinoHistoryWithProvider("redis://localhost:6379/0", provider.TypeRedis, true, "debug")

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
		log.Fatalf("init config failed, err=%v", err)
	}

	var convID = uuid.NewString() // 模拟一个会话id
	cm := createLLMChatModel(ctx)
	// 模拟用户连续问问题
	messList := []string{
		"我数学不好",
		"数学不好可以编程么",
		"今天深圳天气怎么样？",
		"刚才我说我什么科目学得不好来着？请思考之前的对话",
	}

	// 演示使用不同的存储后端
	log.Println("Using MySQL storage:")
	processConversation(ctx, cm, convID, messList, ehMySQL)

	// 使用 Redis 作为存储后端
	convID = uuid.NewString() // 新的会话 ID
	log.Println("\nUsing Redis storage:")
	processConversation(ctx, cm, convID, messList, ehRedis)

	// 演示附件功能
	log.Println("\nAttachment demonstration:")
	attachmentExample()
}

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

// generateResponse uses the chat model to generate a response
func generateResponse(ctx context.Context, cm model.ChatModel, messages []schema.Message) *schema.Message {
	// Convert []schema.Message to []*schema.Message for the Generate method
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
