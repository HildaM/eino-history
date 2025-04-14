package provider

import (
	"fmt"

	"github.com/hildam/eino-history/store/common/logger"
	"github.com/hildam/eino-history/store/mysql"
	"github.com/hildam/eino-history/store/redis"
)

// CreateProvider 创建数据库提供者实例
// 根据传入的配置创建相应类型的数据库提供者实例。如果未指定类型，默认使用MySQL。
// 参数:
//   - config: 数据库配置，包含DSN、类型、是否调试和日志级别
//
// 返回:
//   - Provider: 创建的数据库提供者实例
//   - error: 如果创建过程中发生错误
func CreateProvider(config *Config) (Provider, error) {
	// 如果未指定类型，默认使用MySQL
	if config.Type == "" {
		config.Type = TypeMySQL
	}

	// 如果未指定日志级别，默认不设置（让各Provider自行处理）
	if config.LogLevel == "" {
		config.LogLevel = logger.DefaultLogLevel
	}

	switch config.Type {
	case TypeMySQL:
		return mysql.NewProvider(config.DSN, config.Debug, config.LogLevel)
	case TypeRedis:
		return redis.NewProvider(config.DSN, config.Debug, config.LogLevel)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}
}
