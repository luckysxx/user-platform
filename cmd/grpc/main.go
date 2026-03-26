package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/metrics"
	commonOtel "github.com/luckysxx/common/otel"
	commonRedis "github.com/luckysxx/common/redis"
	"github.com/luckysxx/common/ratelimiter"
	"github.com/luckysxx/common/rpc"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/platform/config"
	"github.com/luckysxx/user-platform/internal/platform/database"
	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/luckysxx/user-platform/internal/service"
	transportgrpc "github.com/luckysxx/user-platform/internal/transport/grpc"

	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger("user-grpc")
	defer log.Sync()

	cfg := config.LoadConfig()

	// 初始化底层基础设施
	entClient, redisClient, publisher := initInfra(cfg, log)
	defer entClient.Close()
	defer redisClient.Close()
	defer publisher.Close()

	// 初始化 OpenTelemetry 链路追踪
	otelShutdown, err := commonOtel.InitTracer(cfg.OTel.ServiceName, cfg.OTel.JaegerEndpoint)
	if err != nil {
		log.Fatal("初始化 OpenTelemetry 失败", zap.Error(err))
	}
	defer otelShutdown(context.Background())

	// 依赖注入与组件装配
	userSvc, authSvc, jwtManager := buildServices(cfg, entClient, redisClient, publisher, log)
	grpcServer := transportgrpc.SetupServer(userSvc, authSvc, jwtManager, log)

	// 启动 Prometheus Metrics HTTP 端点
	go metrics.ServeMetrics(":" + cfg.Metrics.Port)
	log.Info("Metrics 服务已启动", zap.String("port", cfg.Metrics.Port))

	// 阻塞运行与优雅停机
	runServer(grpcServer, cfg.GRPCServer.Port, log)
}

// initInfra 初始化基础设施
func initInfra(cfg *config.Config, log *zap.Logger) (*ent.Client, *redis.Client, event.Publisher) {
	if err := rpc.InitIDGenClient(cfg.IDGenerator.Addr); err != nil {
		log.Fatal("初始化 ID 生成器客户端失败", zap.Error(err))
	}

	entClient := database.InitEntClient(cfg.Database, log)
	redisClient := commonRedis.Init(commonRedis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, log)
	publisher := event.NewKafkaPublisher(cfg.Kafka.Brokers, cfg.Kafka.TopicUserRegistered, log)

	return entClient, redisClient, publisher
}

// buildServices 构建业务层依赖，返回 gRPC Server 所需的 Service 和 JWTManager
func buildServices(cfg *config.Config, entClient *ent.Client, redisClient *redis.Client, publisher event.Publisher, log *zap.Logger) (service.UserService, service.AuthService, *auth.JWTManager) {
	// Repositories
	userRepo := repository.NewUserRepository(entClient)
	outboxRepo := repository.NewEventOutboxRepository(entClient, redisClient)
	tm := repository.NewTransactionManager(entClient)
	sessionRepo := repository.NewRedisSessionRepo(redisClient)
	appRepo := repository.NewAppRepository(entClient)

	// Domain Services
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	rateLim := ratelimiter.NewRedisLimiter(redisClient, log)
	userSvc := service.NewUserService(tm, userRepo, outboxRepo, publisher, log, cfg.Kafka.TopicUserRegistered)
	authSvc := service.NewAuthService(userRepo, appRepo, sessionRepo, jwtManager, rateLim, log)

	return userSvc, authSvc, jwtManager
}

// runServer 启动 gRPC 服务器，监听停机信号后优雅退出
func runServer(s *grpc.Server, port string, log *zap.Logger) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("gRPC 端口监听失败", zap.Error(err))
	}

	go func() {
		log.Info("gRPC 服务已启动", zap.String("port", port))
		if err := s.Serve(lis); err != nil {
			log.Fatal("gRPC 服务异常终止", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("收到停机信号，开始优雅退出...")
	s.GracefulStop()
	log.Info("gRPC 服务已安全退出")
}
