package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hildam/eino-history/model"
	"github.com/hildam/eino-history/store/provider"
)

// AttachmentExample 演示附件功能
func AttachmentExample() {
	log.Println("===== 附件功能示例 =====")

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

	log.Println("===== 附件功能示例结束 =====")
}
