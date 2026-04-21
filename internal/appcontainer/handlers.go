package appcontainer

import (
	"github.com/luckysxx/user-platform/internal/transport/http/server/handler"
	"go.uber.org/zap"
)

func buildUserHandler(services serviceSet, log *zap.Logger) *handler.UserHandler {
	return handler.NewUserHandler(handler.Dependencies{
		UserService: services.userService,
		AuthService: services.authService,
		Logger:      log,
	})
}
