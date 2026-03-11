package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/luckysxx/user-platform/common/auth"
	"github.com/luckysxx/user-platform/common/config"
	"github.com/luckysxx/user-platform/common/database"
	"github.com/luckysxx/user-platform/common/logger"
	"github.com/luckysxx/user-platform/db"
	"github.com/luckysxx/user-platform/handler"
	"github.com/luckysxx/user-platform/repository"
	"github.com/luckysxx/user-platform/service"
	httptransport "github.com/luckysxx/user-platform/transport/http"
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
	conn := database.InitPostgres(cfg.Database, log)
	queries := db.New(conn)

	userRepo := repository.NewUserRepository(queries)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	userSvc := service.NewUserService(userRepo, jwtManager, log)
	userHandler := handler.NewUserHandler(userSvc, log)

	r := gin.New()
	httptransport.SetupRouter(r, userHandler, log)
	r.Run(":" + cfg.Server.Port)
}
