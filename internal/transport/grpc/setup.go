// Package transportgrpc 组装 gRPC Server，对标 HTTP 层的 router.go。
// 职责：注册拦截器链 + 注册 Protobuf 服务。
package transportgrpc

import (
	"github.com/luckysxx/common/metrics"
	auth_pb "github.com/luckysxx/common/proto/auth"
	user_pb "github.com/luckysxx/common/proto/user"
	"github.com/luckysxx/user-platform/internal/service"
	"github.com/luckysxx/user-platform/internal/transport/grpc/interceptor"
	grpcserver "github.com/luckysxx/user-platform/internal/transport/grpc/server"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// SetupServer 组装 gRPC Server（对标 HTTP 的 SetupRouter）
func SetupServer(
	userSvc service.UserService,
	profileSvc service.ProfileService,
	authSvc service.AuthService,
	log *zap.Logger,
) *grpc.Server {
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			metrics.GRPCMetricsInterceptor(),
			interceptor.RecoveryInterceptor(log),
			interceptor.LoggerInterceptor(log),
			interceptor.GatewayAuthInterceptor(),
		),
	)

	// 注册 Protobuf 服务
	user_pb.RegisterUserServiceServer(s, grpcserver.NewUserServer(userSvc, profileSvc, log))
	auth_pb.RegisterAuthServiceServer(s, grpcserver.NewAuthServer(authSvc, log))

	return s
}

