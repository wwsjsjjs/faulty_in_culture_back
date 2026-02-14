package config

import (
	"faulty_in_culture/go_back/internal/infra/logger"
	"os"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	Name       string `yaml:"name"`
	Charset    string `yaml:"charset"`
	AutoCreate bool   `yaml:"auto_create"` // 自动创建数据库
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type ServerConfig struct {
	PublicBaseURL string `yaml:"public_base_url"`
}

// App 应用配置
type App struct {
	Environment string `yaml:"environment"` // 环境：development/production
	Port        string `yaml:"port"`        // 服务端口
	LogMode     string `yaml:"log_mode"`    // 日志模式：dev/prod
	GinMode     string `yaml:"gin_mode"`    // Gin模式：debug/release
}

type MessageConfig struct {
	DelaySeconds        int `yaml:"delay_seconds"`         // 消息延迟处理时间（秒）
	CleanupDays         int `yaml:"cleanup_days"`          // 清理多少天前的已完成消息
	FailedCleanupDays   int `yaml:"failed_cleanup_days"`   // 清理多少天前的失败消息
	CleanupScheduleHour int `yaml:"cleanup_schedule_hour"` // 每天几点执行清理
}

type AIConfig struct {
	APIKey  string `yaml:"api_key"`  // AI API密钥
	BaseURL string `yaml:"base_url"` // AI API基础URL
	Model   string `yaml:"model"`    // 使用的模型
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`       // JWT 签名密钥
	ExpireHours int    `yaml:"expire_hours"` // Token 过期时间（小时）
}

// Config 应用总配置
type Config struct {
	App      App           `yaml:"app"`
	Database DBConfig      `yaml:"database"`
	Redis    RedisConfig   `yaml:"redis"`
	Message  MessageConfig `yaml:"message"`
	Server   ServerConfig  `yaml:"server"`
	AI       AIConfig      `yaml:"ai"`
	JWT      JWTConfig     `yaml:"jwt"`
}

// GlobalConfig 全局配置实例
var GlobalConfig Config

// LoadConfig 加载配置文件
func LoadConfig(path string) {
	f, err := os.Open(path)
	if err != nil {
		logger.Error("failed to open config file", zap.Error(err))
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&GlobalConfig); err != nil {
		logger.Error("failed to decode config file", zap.Error(err))
	}

	// 使用环境变量覆盖配置（Docker部署时优先使用环境变量）
	overrideWithEnv()
}

// overrideWithEnv 使用环境变量覆盖配置
// 优先级：环境变量 > YAML配置文件
func overrideWithEnv() {
	// 数据库配置
	if v := os.Getenv("DB_HOST"); v != "" {
		GlobalConfig.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		GlobalConfig.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		GlobalConfig.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		GlobalConfig.Database.Name = v
	}

	// Redis配置
	if v := os.Getenv("REDIS_HOST"); v != "" {
		GlobalConfig.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		GlobalConfig.Redis.Password = v
	}

	// JWT配置
	if v := os.Getenv("JWT_SECRET"); v != "" {
		GlobalConfig.JWT.Secret = v
	}

	// AI配置
	if v := os.Getenv("HUNYUAN_API_KEY"); v != "" {
		GlobalConfig.AI.APIKey = v
	}

	// 应用配置
	if v := os.Getenv("APP_ENV"); v != "" {
		GlobalConfig.App.Environment = v
	}
	if v := os.Getenv("GIN_MODE"); v != "" {
		GlobalConfig.App.GinMode = v
	}
	if v := os.Getenv("LOG_MODE"); v != "" {
		GlobalConfig.App.LogMode = v
	}

	logger.Info("配置加载完成",
		zap.String("environment", GlobalConfig.App.Environment),
		zap.String("db_host", GlobalConfig.Database.Host),
		zap.String("redis_host", GlobalConfig.Redis.Host))
}
