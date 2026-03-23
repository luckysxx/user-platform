package interceptor

import (
	"context"
	"strings"

	"github.com/luckysxx/user-platform/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

// 定义存储在 Context 中的 Key
const userIDKey contextKey = "user_id"

var authWhiteList = map[string]bool{
	"/user.AuthService/Login":        true,
	"/user.AuthService/RefreshToken": true,
	"/user.UserService/Register":     true,
	"/user.AuthService/VerifyToken":  true,
}


// AuthInterceptor 返回一个用于验证 JWT 并将 UserID 注入上下文的全局拦截器
func AuthInterceptor(jwtManager *auth.JWTManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 1. 白名单配置：如果是登录、刷新 Token 等公开接口，直接放行
		if _, ok := authWhiteList[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		// 2. 核心鉴权逻辑：从上下文中提取 Metadata （也就是客户端传过来的 Header）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "未找到 metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "未提供 Authorization header")
		}

		// Authorization header 的格式通常要求是 "Bearer <token>"
		authHeader := authHeaders[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, status.Error(codes.Unauthenticated, "Authorization header 格式无效")
		}

		tokenString := parts[1]

		// 3. 验证 Token
		claims, err := jwtManager.VerifyToken(tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Token 已过期或无效: %v", err)
		}

		// 4. 将提取出的 UserID 塞到一个全新的 Context 里面向下传
		newCtx := context.WithValue(ctx, userIDKey, claims.UserID)

		return handler(newCtx, req)
	}
}

// UserIDFromContext 是一个辅助函数，留给具体的后端业务服务（比如 auth_server.go）调用取用 UserID
func UserIDFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, status.Error(codes.Unauthenticated, "上下文中不存在 UserID")
	}
	userID, ok := val.(int64)
	if !ok {
		return 0, status.Error(codes.Internal, "上下文中的 UserID 类型错误")
	}
	return userID, nil
}
