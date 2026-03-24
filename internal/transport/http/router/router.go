package router

import (
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/transport/http/handler"
	"github.com/luckysxx/user-platform/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(r *gin.Engine, userHandler *handler.UserHandler, jwtManager *auth.JWTManager, log *zap.Logger) {
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.GinLogger(log))
	r.Use(middleware.GinRecovery(log, true))

	// 用于 Docker 容器的健康检查
	healthHandler := func(c *gin.Context) {
		c.String(200, "ok")
	}
	r.GET("/health", healthHandler)
	r.HEAD("/health", healthHandler)

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.POST("/refresh", userHandler.RefreshToken)

			// 需要鉴权接口组 (退化为信任网关传递的信息)
			authUsers := users.Group("")
			authUsers.Use(middleware.GatewayAuth(log))
			{
				authUsers.POST("/logout", userHandler.Logout)
			}
		}
	}
}
