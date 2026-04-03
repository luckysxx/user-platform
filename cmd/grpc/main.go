package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/luckysxx/common/health"
	"github.com/luckysxx/common/logger"
	commonOtel "github.com/luckysxx/common/otel"
	"github.com/luckysxx/common/ratelimiter"
	commonRedis "github.com/luckysxx/common/redis"
	"github.com/luckysxx/common/rpc"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/platform/bootstrap"
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
	entClient, redisClient := initInfra(cfg, log)
	defer entClient.Close()
	defer redisClient.Close()

	// 初始化 OpenTelemetry 链路追踪
	otelShutdown, err := commonOtel.InitTracer(cfg.OTel.ServiceName, cfg.OTel.JaegerEndpoint)
	if err != nil {
		log.Fatal("初始化 OpenTelemetry 失败", zap.Error(err))
	}
	defer otelShutdown(context.Background())

	// 依赖注入与组件装配
	healthChecker := buildHealthChecker(entClient, redisClient)
	grpcHealthServer := grpchealth.NewServer()
	startHealthSync(healthChecker, grpcHealthServer, log)
	userSvc, profileSvc, authSvc, _ := buildServices(cfg, entClient, redisClient, log)
	grpcServer := transportgrpc.SetupServer(userSvc, profileSvc, authSvc, grpcHealthServer, log)
	adminServer := buildAdminServer(cfg, healthChecker)

	// 阻塞运行与优雅停机
	runServer(grpcServer, grpcHealthServer, adminServer, cfg.GRPCServer.Port, log)
}

// initInfra 初始化基础设施
func initInfra(cfg *config.Config, log *zap.Logger) (*ent.Client, *redis.Client) {
	if err := rpc.InitIDGenClient(cfg.IDGenerator.Addr); err != nil {
		log.Fatal("初始化 ID 生成器客户端失败", zap.Error(err))
	}

	entClient := database.InitEntClient(cfg.Database.Driver, cfg.Database.Source, cfg.Database.AutoMigrate, log)
	if err := bootstrap.EnsureDefaultApps(context.Background(), entClient, log, bootstrap.DefaultApps); err != nil {
		log.Fatal("初始化默认应用失败", zap.Error(err))
	}
	redisClient := commonRedis.Init(commonRedis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, log)

	return entClient, redisClient
}

// buildServices 构建业务层依赖，返回 gRPC Server 所需的 Service 和 JWTManager
func buildServices(cfg *config.Config, entClient *ent.Client, redisClient *redis.Client, log *zap.Logger) (service.UserService, service.ProfileService, service.AuthService, *auth.JWTManager) {
	// Repositories
	userRepo := repository.NewUserRepository(entClient)
	profileRepo := repository.NewProfileRepository(entClient)
	outboxRepo := repository.NewEventOutboxRepository(entClient)
	tm := repository.NewTransactionManager(entClient)
	sessionRepo := repository.NewRedisSessionRepo(redisClient)
	appRepo := repository.NewAppRepository(entClient)

	// Domain Services
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	rateLim := ratelimiter.NewFixedWindowLimiter(redisClient, log)
	userSvc := service.NewUserService(tm, userRepo, profileRepo, outboxRepo, log, cfg.Kafka.TopicUserRegistered)
	profileSvc := service.NewProfileService(profileRepo, log)
	authSvc := service.NewAuthService(userRepo, appRepo, sessionRepo, jwtManager, rateLim, log)

	return userSvc, profileSvc, authSvc, jwtManager
}

func buildHealthChecker(entClient *ent.Client, redisClient *redis.Client) *health.Checker {
	healthChecker := health.NewChecker()
	healthChecker.AddCheck("postgres", func(ctx context.Context) error {
		_, err := entClient.User.Query().Exist(ctx)
		return err
	})
	healthChecker.AddCheck("redis", func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	})
	return healthChecker
}

// buildAdminServer 为纯 gRPC 进程构建一个旁路 HTTP 管理端口，统一暴露 metrics 和健康探针。
func buildAdminServer(cfg *config.Config, healthChecker *health.Checker) *http.Server {

	mux := http.NewServeMux()
	healthChecker.RegisterHTTP(mux)
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:    ":" + cfg.Metrics.Port,
		Handler: mux,
	}
}

// startHealthSync 将依赖检查结果同步到 gRPC 原生 Health 服务。
func startHealthSync(checker *health.Checker, healthServer *grpchealth.Server, log *zap.Logger) {
	var lastStatus healthgrpc.HealthCheckResponse_ServingStatus
	var initialized bool

	update := func() {
		allHealthy, results := checker.Evaluate(context.Background())
		status := healthgrpc.HealthCheckResponse_SERVING
		if !allHealthy {
			status = healthgrpc.HealthCheckResponse_NOT_SERVING
		}

		healthServer.SetServingStatus("", status)
		healthServer.SetServingStatus("user.UserService", status)
		healthServer.SetServingStatus("user.AuthService", status)

		if initialized && status == lastStatus {
			return
		}
		lastStatus = status
		initialized = true

		if allHealthy {
			log.Debug("gRPC health 状态已更新", zap.String("status", status.String()))
			return
		}

		log.Warn("gRPC health 状态已更新",
			zap.String("status", status.String()),
			zap.Any("checks", results),
		)
	}

	update()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			update()
		}
	}()
}

// runServer 启动 gRPC 服务器，监听停机信号后优雅退出
func runServer(s *grpc.Server, healthServer *grpchealth.Server, adminServer *http.Server, port string, log *zap.Logger) {
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

	go func() {
		log.Info("gRPC 管理端口已启动",
			zap.String("port", adminServer.Addr),
			zap.String("endpoints", "/metrics, /healthz, /readyz"),
		)
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("gRPC 管理端口异常终止", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("收到停机信号，开始优雅退出...")
	healthServer.SetServingStatus("", healthgrpc.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus("user.UserService", healthgrpc.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus("user.AuthService", healthgrpc.HealthCheckResponse_NOT_SERVING)
	s.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := adminServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal("gRPC 管理端口强制退出", zap.Error(err))
	}

	log.Info("gRPC 服务已安全退出")
}
