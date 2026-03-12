package middleware

import (
	"errors"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 认证中间件
func JWTAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := auth.AuthenticateBearerToken(jwtManager, c.GetHeader("Authorization"))
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrMissingAuthHeader):
				response.Unauthorized(c, "未携带 Token")
			case errors.Is(err, auth.ErrInvalidAuthHeaderFormat):
				response.Unauthorized(c, "Token 格式错误")
			default:
				response.Unauthorized(c, "Token 无效或已过期")
			}
			c.Abort()
			return
		}

		// 将 UserID 存入上下文，供后续业务逻辑使用
		c.Set("userID", userID)

		c.Next() // 放行
	}
}
