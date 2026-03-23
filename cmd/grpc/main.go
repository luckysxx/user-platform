package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/redis/go-redis/v9"
	"github.com/luckysxx/common/logger"
	auth_pb "github.com/luckysxx/common/proto/auth"
	user_pb "github.com/luckysxx/common/proto/user"
	"github.com/luckysxx/common/ratelimiter"
	"github.com/luckysxx/common/rpc"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/platform/config"
	"github.com/luckysxx/user-platform/internal/platform/database"
	"github.com/luckysxx/user-platform/internal/cache"
	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/luckysxx/user-platform/internal/service"
	usergrpcserver "github.com/luckysxx/user-platform/internal/transport/grpc/server"
	"github.com/luckysxx/user-platform/internal/transport/grpc/interceptor"

	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger("user-grpc")
	defer log.Sync()

	cfg := config.LoadConfig()

	// 1. 初始化底层基础设施 (Infrastructure Bootstrapping)
	entClient, redisClient, publisher := initInfra(cfg, log)
	defer entClient.Close()
	defer redisClient.Close()
	defer publisher.Close()

	// 2. 依赖注入与组件装配 (Dependency Injection)
	grpcServer := buildServer(cfg, entClient, redisClient, publisher, log)

	// 3. 阻塞运行与优雅停机 (Graceful Shutdown)
	runServer(grpcServer, cfg.GRPCServer.Port, log)
}

// initInfra 抽取基础设施的初始化逻辑
func initInfra(cfg *config.Config, log *zap.Logger) (*ent.Client, *redis.Client, event.Publisher) {
	if err := rpc.InitIDGenClient(cfg.IDGenerator.Addr); err != nil {
		log.Fatal("初始化 ID 生成器客户端失败", zap.Error(err))
	}

	entClient := database.InitEntClient(cfg.Database, log)
	redisClient := cache.InitRedis(cfg.Redis, log)
	publisher := event.NewKafkaPublisher(cfg.Kafka.Brokers, log)

	return entClient, redisClient, publisher
}

// buildServer 抽取所有的依赖注入逻辑
func buildServer(cfg *config.Config, entClient *ent.Client, redisClient *redis.Client, publisher event.Publisher, log *zap.Logger) *grpc.Server {
	// Repositories
	userRepo := repository.NewUserRepository(entClient)
	outboxRepo := repository.NewEventOutboxRepository(entClient)
	tm := repository.NewTransactionManager(entClient)
	sessionRepo := repository.NewRedisSessionRepo(redisClient)
	appRepo := repository.NewAppRepository(entClient)

	// Domain Services
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	rateLim := ratelimiter.NewRedisLimiter(redisClient, log)
	userSvc := service.NewUserService(tm, userRepo, outboxRepo, publisher, log)
	authSvc := service.NewAuthService(userRepo, appRepo, sessionRepo, jwtManager, rateLim, log)

	// GRPC Handlers
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(log),
			interceptor.LoggerInterceptor(log),
			interceptor.AuthInterceptor(jwtManager),
		),
	)
	user_pb.RegisterUserServiceServer(s, usergrpcserver.NewUserServer(userSvc, log))
	auth_pb.RegisterAuthServiceServer(s, usergrpcserver.NewAuthServer(authSvc, log))

	return s
}

// runServer 抽取 GRPC 服务器的启动与停机监听
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
