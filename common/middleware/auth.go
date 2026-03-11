package middleware

import (
	"strings"

	"github.com/luckysxx/user-platform/common/auth"
	"github.com/luckysxx/user-platform/common/response"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 认证中间件
func JWTAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Header 中的 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未携带 Token")
			c.Abort() // 阻止后续处理
			return
		}

		// 2. 格式通常是 "Bearer xxxxxxx"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Token 格式错误")
			c.Abort()
			return
		}

		// 3. 解析 Token
		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 4. 将 UserID 存入上下文，供后续业务逻辑使用
		c.Set("userID", claims.UserID)

		c.Next() // 放行
	}
}
