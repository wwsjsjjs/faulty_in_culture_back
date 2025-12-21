package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitLogger 初始化日志系统
// mode: "dev" 开发模式（输出到终端），"prod" 生产模式（输出到文件）
func InitLogger(mode string) error {
	var core zapcore.Core

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色输出
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if mode == "dev" {
		// 开发模式：输出到终端
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
	} else {
		// 生产模式：输出到文件
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 无彩色
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

		// 配置日志轮转
		logFile := &lumberjack.Logger{
			Filename:   filepath.Join("logs", "app.log"),
			MaxSize:    100,  // 单个文件最大 100MB
			MaxBackups: 10,   // 保留最多 10 个备份
			MaxAge:     30,   // 保留最多 30 天
			Compress:   true, // 压缩备份文件
		}

		core = zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(logFile),
			zapcore.InfoLevel,
		)
	}

	// 创建 logger
	Logger = zap.New(core,
		zap.AddCaller(),                       // 添加调用者信息
		zap.AddCallerSkip(1),                  // 跳过一层调用栈
		zap.AddStacktrace(zapcore.ErrorLevel), // Error 级别及以上记录堆栈
	)

	return nil
}

// customTimeEncoder 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

// Info 记录 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 记录 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 记录 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Debug 记录 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Sync 刷新日志缓冲
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
