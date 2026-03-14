package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/cache"
	"github.com/luckysxx/user-platform/internal/platform/config"
	"github.com/luckysxx/user-platform/internal/platform/database"
	"github.com/luckysxx/user-platform/internal/platform/logger"
	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/luckysxx/user-platform/internal/service"
	usergrpcserver "github.com/luckysxx/user-platform/internal/transport/grpc/server"
	pb "github.com/luckysxx/user-platform/proto/user"

	"go.uber.org/zap"
)

func main() {
	port := os.Getenv("USER_GRPC_PORT")
	if port == "" {
		port = "9091"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logg := logger.NewLogger("user-grpc")
		defer logg.Sync()
		logg.Fatal("grpc listen failed", zap.Error(err), zap.String("port", port))
	}

	logg := logger.NewLogger("user-grpc")
	defer logg.Sync()
	cfg := config.LoadConfig()
	entClient := database.InitEntClient(cfg.Database, logg)
	defer entClient.Close()
	redisClient := cache.InitRedis(cfg.Redis, logg)
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(entClient)
	userSvc := service.NewUserService(userRepo, logg)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	authSvc := service.NewAuthService(userRepo, redisClient, jwtManager, logg)

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, usergrpcserver.NewUserServer(userSvc, logg))
	pb.RegisterAuthServiceServer(s, usergrpcserver.NewAuthServer(authSvc, logg))
	logg.Info("user grpc listening", zap.String("port", port))

	go func() {
		if err := s.Serve(lis); err != nil {
			logg.Fatal("grpc serve failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logg.Info("shutting down user grpc server")
	s.GracefulStop()
	logg.Info("user grpc server stopped")
}
