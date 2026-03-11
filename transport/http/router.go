package http

import (
	"github.com/luckysxx/user-platform/common/middleware"
	"github.com/luckysxx/user-platform/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(r *gin.Engine, userHandler *handler.UserHandler, log *zap.Logger) {
	r.Use(middleware.GinLogger(log))
	r.Use(middleware.GinRecovery(log, true))

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
		}
	}
}
