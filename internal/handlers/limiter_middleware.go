package handlers

import (
	"faulty_in_culture/go_back/internal/logger"

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

func InitLimiters() {
	// Redis client (v9)
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	store, err := redisStore.NewStoreWithOptions(client, limiter.StoreOptions{})
	if err != nil {
		logger.Error("限流器初始化失败", zap.Error(err))
		panic(err)
	}

	// 全局限流参数 60次/分钟（如需调整只改这里）
	rateGlobal, _ := limiter.NewRateFromFormatted("60-M")
	LimiterGlobal = limiterGin.NewMiddleware(limiter.New(store, rateGlobal))
}
