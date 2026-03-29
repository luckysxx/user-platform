package interceptor

import (
	"context"
	"time"

	"github.com/luckysxx/common/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggerInterceptor 返回一个一元服务器拦截器，用于记录 gRPC 方法的调用耗时和执行结果。
// 自动从 context 中提取 OTel TraceID 注入到日志中。
func LoggerInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var userID int64
		if val := ctx.Value(userIDKey); val != nil {
			if uid, ok := val.(int64); ok {
				userID = uid
			}
		}
		start := time.Now()

		// 继续执行实际的 Handler
		resp, err = handler(ctx, req)

		duration := time.Since(start)

		// 从 context 提取 TraceID，自动附加到日志
		reqLog := logger.Ctx(ctx, log)

		// 根据错误类型记录不同级别的日志
		if err != nil {
			st, _ := status.FromError(err)
			code := st.Code()
			// 客户端原因导致的错误记录为 Warn，服务端异常（Internal, Unavailable等）记录为 Error
			if code == codes.Canceled || code == codes.DeadlineExceeded || code == codes.InvalidArgument || code == codes.Unauthenticated || code == codes.PermissionDenied || code == codes.NotFound {
				reqLog.Warn("gRPC 请求失败",
					zap.String("method", info.FullMethod),
					zap.Int64("user_id", userID),
					zap.Duration("duration", duration),
					zap.String("code", code.String()),
					zap.Error(err),
				)
			} else {
				reqLog.Error("gRPC 请求失败",
					zap.String("method", info.FullMethod),
					zap.Int64("user_id", userID),
					zap.Duration("duration", duration),
					zap.String("code", st.Code().String()),
					zap.Error(err),
				)
			}
		} else {
			reqLog.Info("gRPC 请求成功",
				zap.String("method", info.FullMethod),
				zap.Int64("user_id", userID),
				zap.Duration("duration", duration),
				zap.String("code", codes.OK.String()),
			)
		}

		return resp, err
	}
}

