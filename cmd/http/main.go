package main

import (
	"github.com/gin-gonic/gin"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/cache"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/platform/config"
	"github.com/luckysxx/user-platform/internal/platform/database"
	"github.com/luckysxx/user-platform/internal/platform/logger"
	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/luckysxx/user-platform/internal/service"
	"github.com/luckysxx/user-platform/internal/transport/http/handler"
	httprouter "github.com/luckysxx/user-platform/internal/transport/http/router"
)

// @title           User Platform Service
// @version         1.0
// @description     用户中心服务，提供注册、登录功能
// @host            localhost:8081
// @BasePath        /api/v1
func main() {
	log := logger.NewLogger("user")
	defer log.Sync()

	cfg := config.LoadConfig()
	entClient := database.InitEntClient(cfg.Database, log)
	defer entClient.Close()
	redisClient := cache.InitRedis(cfg.Redis, log)
	defer redisClient.Close()

	publisher := event.NewKafkaPublisher(cfg.Kafka.Brokers, log)
	defer publisher.Close()
	userRepo := repository.NewUserRepository(entClient)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	userSvc := service.NewUserService(userRepo, publisher, log)
	authSvc := service.NewAuthService(userRepo, redisClient, jwtManager, log)
	userHandler := handler.NewUserHandler(userSvc, authSvc, log)

	r := gin.New()
	httprouter.SetupRouter(r, userHandler, log)
	r.Run(":" + cfg.Server.Port)
}
