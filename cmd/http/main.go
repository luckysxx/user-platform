package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/ratelimiter"
	"github.com/luckysxx/common/rpc"
	commonRedis "github.com/luckysxx/common/pkg/redis"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/platform/config"
	"github.com/luckysxx/user-platform/internal/platform/database"
	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/luckysxx/user-platform/internal/service"
	"github.com/luckysxx/user-platform/internal/transport/http/handler"
	httprouter "github.com/luckysxx/user-platform/internal/transport/http/router"
	"github.com/luckysxx/user-platform/internal/worker"
	"go.uber.org/zap"
)

// @title           User Platform Service
// @version         1.0
// @description     用户中心服务，提供注册、登录功能
// @host            localhost:8081
// @BasePath        /api/v1
func main() {
	log := logger.NewLogger("user-http")
	defer log.Sync()

	cfg := config.LoadConfig()

	// 1. 初始化底层基础设施
	entClient, redisClient, kafkaWriter := initInfra(cfg, log)
	defer entClient.Close()
	defer redisClient.Close()
	defer kafkaWriter.Close()

	// 2. 依赖注入与组件装配
	publisher := event.NewKafkaPublisher(cfg.Kafka.Brokers, log)
	defer publisher.Close()
	router := buildRouter(cfg, entClient, redisClient, publisher, log)

	// 3. 启动 OutboxWorker（异步补偿发件箱）
	outboxRepo := repository.NewEventOutboxRepository(entClient)
	outboxWorker := worker.NewOutboxWorker(outboxRepo, kafkaWriter, log)

	// 4. 阻塞运行与优雅停机
	runServer(router, outboxWorker, cfg.Server.Port, log)
}

// initInfra 初始化基础设施
func initInfra(cfg *config.Config, log *zap.Logger) (*ent.Client, *redis.Client, *kafka.Writer) {
	if err := rpc.InitIDGenClient(cfg.IDGenerator.Addr); err != nil {
		log.Fatal("初始化 ID 生成器客户端失败", zap.Error(err))
	}

	entClient := database.InitEntClient(cfg.Database, log)
	redisClient := commonRedis.Init(commonRedis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, log)

	kafkaWriter := event.NewKafkaWriter(cfg.Kafka.Brokers)

	return entClient, redisClient, kafkaWriter
}

// buildRouter 依赖注入装配
func buildRouter(cfg *config.Config, entClient *ent.Client, redisClient *redis.Client, publisher event.Publisher, log *zap.Logger) *gin.Engine {
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

	// Transport
	userHandler := handler.NewUserHandler(userSvc, authSvc, log)
	r := gin.New()
	httprouter.SetupRouter(r, userHandler, jwtManager, log)

	return r
}

// runServer 启动 HTTP 服务器和 OutboxWorker，监听停机信号后优雅退出
func runServer(router *gin.Engine, outboxWorker *worker.OutboxWorker, port string, log *zap.Logger) {
	// 创建一个全局的可取消 context，用于统一管理所有后台协程的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 HTTP 服务器
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	go func() {
		log.Info("HTTP 服务已启动", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP 服务监听失败", zap.Error(err))
		}
	}()

	// 启动 OutboxWorker（后台轮询协程）
	go func() {
		if err := outboxWorker.Run(ctx); err != nil {
			log.Error("OutboxWorker exited with error", zap.Error(err))
		}
	}()

	// 监听停机信号
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("收到停机信号，开始优雅退出...")

	// 1. 先取消 context，通知 OutboxWorker 停止轮询
	cancel()

	// 2. 优雅关闭 HTTP 服务器（等待正在处理的请求完成）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("HTTP 服务强制退出", zap.Error(err))
	}

	log.Info("所有服务已安全退出")
}
