// Package transportgrpc 组装 gRPC Server，对标 HTTP 层的 router.go。
// 职责：注册拦截器链 + 注册 Protobuf 服务。
package transportgrpc

import (
	"time"

	commonlogger "github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/metrics"
	auth_pb "github.com/luckysxx/common/proto/auth"
	user_pb "github.com/luckysxx/common/proto/user"
	"github.com/luckysxx/user-platform/internal/service"
	"github.com/luckysxx/user-platform/internal/transport/grpc/interceptor"
	grpcserver "github.com/luckysxx/user-platform/internal/transport/grpc/server"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// SetupServer 组装 gRPC Server（对标 HTTP 的 SetupRouter）
func SetupServer(
	userSvc service.UserService,
	profileSvc service.ProfileService,
	authSvc service.AuthService,
	healthServer healthgrpc.HealthServer,
	log *zap.Logger,
) *grpc.Server {
	s := grpc.NewServer(
		// ── Keepalive ────────────────────────────────────────
		// 放宽心跳限制，允许 API-Gateway 以 10s 间隔发送 ping。
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // 允许客户端最小 5 秒 ping 一次
			PermitWithoutStream: true,            // 允许无活跃流时的 ping
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,  // 空闲连接 5 分钟后关闭
			MaxConnectionAge:      30 * time.Minute, // 连接最长存活 30 分钟，促进负载均衡
			MaxConnectionAgeGrace: 10 * time.Second, // 关闭前给 10 秒完成正在执行的 RPC
			Time:                  15 * time.Second,  // 服务端每 15 秒向客户端发 ping
			Timeout:               5 * time.Second,   // 5 秒没响应视为断连
		}),
		// ── Observability ────────────────────────────────────
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			metrics.GRPCMetricsInterceptor(),
			interceptor.RecoveryInterceptor(log),
			interceptor.GatewayAuthInterceptor(),
			commonlogger.GRPCUnaryServerInterceptor(log, interceptor.LogFieldsFromContext),
		),
	)

	// 注册 Protobuf 服务
	user_pb.RegisterUserServiceServer(s, grpcserver.NewUserServer(userSvc, profileSvc, log))
	auth_pb.RegisterAuthServiceServer(s, grpcserver.NewAuthServer(authSvc, log))

	// 注册 gRPC 原生健康检查服务，供标准 Health/Check RPC 调用。
	healthgrpc.RegisterHealthServer(s, healthServer)

	return s
}
