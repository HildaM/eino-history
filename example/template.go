package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func createTemplate(ctx context.Context) prompt.ChatTemplate {
	// 创建模板，使用 FString 格式
	return prompt.FromMessages(schema.FString,
		// 系统消息模板
		schema.SystemMessage("你是一个{role}。你需要用{style}的语气回答问题。你的目标是帮助程序员保持积极乐观的心态，提供技术建议的同时也要关注他们的心理健康。"),

		// 插入需要的对话历史（新对话的话这里不填）
		schema.MessagesPlaceholder("chat_history", true),

		// 用户消息模板
		schema.UserMessage("问题: {question}"),
	)
}

func createMessagesFromTemplate(ctx context.Context, convID, question string) (messages []*schema.Message, err error) {
	template := createTemplate(ctx)
	/* add start */
	chatHistory, err := ehMySQL.GetHistory(convID, 100)
	if err != nil {
		return
	}
	// 插入一条用户数据
	err = ehMySQL.SaveMessage(&schema.Message{
		Role:    schema.User,
		Content: question,
	}, convID)
	if err != nil {
		return
	}
	/* add end */
	// 使用模板生成消息
	messages, err = template.Format(context.Background(), map[string]any{
		"role":         "程序员鼓励师",
		"style":        "积极、温暖且专业",
		"question":     question, // "我的代码一直报错，感觉好沮丧，该怎么办？",
		"chat_history": chatHistory,
	})
	if err != nil {
		return
	}
	return
}

func generate(ctx context.Context, llm model.ChatModel, in []*schema.Message) *schema.Message {
	result, err := llm.Generate(ctx, in)
	if err != nil {
		log.Fatalf("llm generate failed: %v", err)
	}
	return result
}

// GetQuestionsByTopic 根据主题返回相应的问题列表
func GetQuestionsByTopic(topic string) []string {
	questionMap := map[string][]string{
		"数学学习": {
			"我数学不好，特别是微积分，有什么好的学习方法吗？",
			"如何提高数学解题速度？",
		},
		"编程学习": {
			"编程需要很强的数学基础吗？",
			"初学者应该选择哪种编程语言？",
		},
		"对话回顾": {
			"刚才我们聊了哪些话题？",
			"你能回忆之前我问过的问题吗？",
			"我们之前讨论的主要内容是什么？",
			"我上一次提到了什么科目？请思考之前的对话",
		},
	}

	// 如果找不到对应主题的问题，返回默认问题
	questions, exist := questionMap[topic]
	if !exist {
		return []string{
			"这是一个新主题，能介绍一下吗？",
			"这个主题有哪些要点？",
			"为什么这个主题很重要？",
			"如何深入学习这个主题？",
		}
	}

	return questions
}
