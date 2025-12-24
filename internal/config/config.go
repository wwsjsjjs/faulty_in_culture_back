package config

import (
	"faulty_in_culture/go_back/internal/logger"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Charset  string `yaml:"charset"`
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

type MessageConfig struct {
	DelaySeconds        int `yaml:"delay_seconds"`         // 消息延迟处理时间（秒）
	CleanupDays         int `yaml:"cleanup_days"`          // 清理多少天前的已完成消息
	FailedCleanupDays   int `yaml:"failed_cleanup_days"`   // 清理多少天前的失败消息
	CleanupScheduleHour int `yaml:"cleanup_schedule_hour"` // 每天几点执行清理
}

type Config struct {
	Database DBConfig      `yaml:"database"`
	Redis    RedisConfig   `yaml:"redis"`
	Message  MessageConfig `yaml:"message"`
	Server   ServerConfig  `yaml:"server"`
}

var AppConfig Config

func LoadConfig(path string) {
	f, err := os.Open(path)
	if err != nil {
		logger.Error("failed to open config file", zap.Error(err))
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&AppConfig); err != nil {
		logger.Error("failed to decode config file", zap.Error(err))
	}
}
