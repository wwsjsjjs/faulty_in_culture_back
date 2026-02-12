package middleware

import (
	"faulty_in_culture/go_back/internal/shared/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	limiterGin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
	"go.uber.org/zap"
)

var (
	LimiterGlobal gin.HandlerFunc
)

// InitLimiters 初始化限流器
func InitLimiters(redisAddr, redisPassword string, redisDB int) error {
	logger.Info("middleware.InitLimiters",
		zap.String("redisAddr", redisAddr),
		zap.Int("redisDB", redisDB),
	)

	// Redis client (v9)
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	store, err := redisStore.NewStoreWithOptions(client, limiter.StoreOptions{})
	if err != nil {
		logger.Error("middleware.InitLimiters: 创建 Redis store 失败",
			zap.Error(err),
		)
		return err
	}

	// 全局限流参数 60次/分钟（如需调整只改这里）
	rateGlobal, err := limiter.NewRateFromFormatted("60-M")
	if err != nil {
		logger.Error("middleware.InitLimiters: 创建限流规则失败",
			zap.Error(err),
		)
		return err
	}

	LimiterGlobal = limiterGin.NewMiddleware(limiter.New(store, rateGlobal))

	logger.Info("middleware.InitLimiters: 限流器初始化成功")
	return nil
}
