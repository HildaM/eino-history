# Eino-History

聊天历史记录管理库，为Eino大模型框架提供历史记录管理功能，支持多轮对话。

## 项目说明

本项目是对 [wangle201210/chat-history](https://github.com/wangle201210/chat-history) 的优化与扩展，在原项目的基础上增加了以下功能：

- 支持Redis存储后端，除了MySQL以外的选择
- 增强的附件管理功能，支持图片、文件、代码、音频和视频等多种类型

## 功能特点

- 支持持久化存储聊天历史记录
- 支持MySQL和Redis两种存储后端
- 支持附件管理功能（图片、文件等）
- 支持对话会话管理
- 简单易用的API接口

## 安装

```bash
go get github.com/hildam/eino-history
```

## 快速开始

### 初始化

```go
// 方式1：使用MySQL（默认）
ehMySQL := eino.NewDefaultEinoHistory(
    "root:password@tcp(127.0.0.1:3306)/chat_history?charset=utf8mb4&parseTime=True&loc=Local", 
    "debug"
)

// 方式2：使用Redis
ehRedis := eino.NewEinoHistoryWithProvider(
    "redis://localhost:6379/0", 
    provider.TypeRedis, 
    true, 
    "debug"
)

// 确保最后关闭连接
defer func() {
    if err := ehMySQL.Close(); err != nil {
        log.Printf("关闭 MySQL 连接失败: %v", err)
    }
    if err := ehRedis.Close(); err != nil {
        log.Printf("关闭 Redis 连接失败: %v", err)
    }
}()
```

### 基本使用

```go
// 创建会话ID
convID := uuid.NewString()

// 获取历史记录
chatHistory, err := ehMySQL.GetHistory(convID, 100)
if err != nil {
    log.Fatalf("获取历史记录失败: %v", err)
}

// 保存用户消息
err = ehMySQL.SaveMessage(&schema.Message{
    Role:    schema.User,
    Content: "用户问题",
}, convID)
if err != nil {
    log.Fatalf("保存用户消息失败: %v", err)
}

// 保存AI回复
err = ehMySQL.SaveMessage(&schema.Message{
    Role:    schema.Assistant,
    Content: "AI回复内容",
}, convID)
if err != nil {
    log.Fatalf("保存AI回复失败: %v", err)
}
```

## 集成到Eino框架

以下是将Eino-History集成到现有Eino项目的步骤：

1. 初始化EinoHistory实例：

```go
var eh = eino.NewEinoHistory("root:password@tcp(127.0.0.1:3306)/chat_history")
```

2. 在原来获取messages的函数中添加历史记录相关代码：

```go
func createMessagesFromTemplate(ctx context.Context, convID, question string) (messages []*schema.Message, err error) {
    template := createTemplate(ctx)
    
    // 获取历史记录
    chatHistory, err := eh.GetHistory(convID, 100)
    if err != nil {
        return
    }
    
    // 保存用户消息
    err = eh.SaveMessage(&schema.Message{
        Role:    schema.User,
        Content: question,
    }, convID)
    if err != nil {
        return
    }
    
    // 使用模板生成消息
    messages, err = template.Format(context.Background(), map[string]any{
        "role":         "程序员鼓励师",
        "style":        "积极、温暖且专业",
        "question":     question,
        "chat_history": chatHistory,
    })
    if err != nil {
        return
    }
    
    return
}
```

3. 在处理LLM回复的地方添加保存回复的代码：

```go
messages, err := createMessagesFromTemplate(ctx, convID, userInput)
if err != nil {
    log.Fatalf("create messages failed: %v", err)
    return
}

result := generate(ctx, cm, messages)

// 保存AI回复
err = eh.SaveMessage(result, convID)
if err != nil {
    log.Fatalf("save assistant message err: %v", err)
    return
}
```

## 完整示例

请参考项目中的 [example](./example) 目录，包含完整的使用示例：
- MySQL和Redis两种存储后端的使用
- 多轮对话的处理流程
- 附件功能的使用

## 与原项目的对比

相比于原项目 [wangle201210/chat-history](https://github.com/wangle201210/chat-history)，本项目主要做了以下改进：

1. **存储后端扩展**：增加了Redis支持，使用户可以根据需求选择合适的存储方式
2. **附件功能增强**：完整支持附件的管理，包括图片、文件、代码、音频和视频等
3. **数据模型优化**：优化数据库表结构，增加了更多的索引和字段
4. **易用性改进**：提供了更简洁的API和更详细的示例代码
5. **文档完善**：更全面的使用说明和接口文档

## 附件功能

Eino-History支持管理附件，如图片、文件、代码、音频和视频：

```go
// 获取仓库
attachmentRepo := dbProvider.GetAttachmentStore()
messageAttachmentRepo := dbProvider.GetMessageAttachmentStore()

// 创建附件
attachment := &models.Attachment{
    AttachID:       uuid.NewString(),
    AttachmentType: "image",
    FileName:       "example.jpg",
    FileSize:       1024 * 100, // 100KB
    StorageType:    "path",
    StoragePath:    "/images/example.jpg",
    MimeType:       "image/jpeg",
    CreatedAt:      time.Now().Unix(),
}

// 保存附件
if err := attachmentRepo.Create(attachment); err != nil {
    log.Fatalf("创建附件失败: %v", err)
}

// 关联附件到消息
messageAttachment := &models.MessageAttachment{
    MessageID:    messageID,
    AttachmentID: attachment.AttachID,
}

if err := messageAttachmentRepo.Create(messageAttachment); err != nil {
    log.Fatalf("关联附件到消息失败: %v", err)
}

// 获取消息的所有附件
attachmentList, err := attachmentRepo.ListByMessage(messageID)
if err != nil {
    log.Fatalf("获取消息附件详情失败: %v", err)
}
```

## 配置

配置放在 main.go 同级目录中

项目使用YAML配置文件，示例配置如下：

```yaml
# DeepSeek API配置
DeekSeek:
  # DeepSeek API密钥
  api_key: "your-api-key"

  # 使用的模型ID
  model_id: "deepseek-chat"

  # API基础URL
  base_url: "https://api.deepseek.com/v1"

# 可以添加更多配置项
# 例如：
# server:
#   port: 8080
#   timeout: 30s
```

## 数据库设计

项目包含以下几个核心表：

1. `conversations` - 会话表
2. `messages` - 消息表
3. `attachments` - 附件表
4. `message_attachments` - 消息与附件的关联表

## 贡献

欢迎提交Issue和Pull Request。

## 致谢

特别感谢 [wangle201210/chat-history](https://github.com/wangle201210/chat-history) 项目提供的基础实现，本项目在此基础上进行了功能扩展与优化。

## 许可证

请查看项目中的 [LICENSE](./LICENSE) 文件。 