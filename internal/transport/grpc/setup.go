// Package transportgrpc 组装 gRPC Server，对标 HTTP 层的 router.go。
// 职责：注册拦截器链 + 注册 Protobuf 服务。
package transportgrpc

import (
	"github.com/luckysxx/common/metrics"
	auth_pb "github.com/luckysxx/common/proto/auth"
	user_pb "github.com/luckysxx/common/proto/user"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/service"
	"github.com/luckysxx/user-platform/internal/transport/grpc/interceptor"
	grpcserver "github.com/luckysxx/user-platform/internal/transport/grpc/server"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// SetupServer 组装 gRPC Server（对标 HTTP 的 SetupRouter）
// 集中管理拦截器链和服务注册，保持 main.go 简洁。
func SetupServer(
	userSvc service.UserService,
	authSvc service.AuthService,
	jwtManager *auth.JWTManager,
	log *zap.Logger,
) *grpc.Server {
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			metrics.GRPCMetricsInterceptor(),
			interceptor.RecoveryInterceptor(log),
			interceptor.LoggerInterceptor(log),
			interceptor.AuthInterceptor(jwtManager),
		),
	)

	// 注册 Protobuf 服务
	user_pb.RegisterUserServiceServer(s, grpcserver.NewUserServer(userSvc, log))
	auth_pb.RegisterAuthServiceServer(s, grpcserver.NewAuthServer(authSvc, log))

	return s
}
