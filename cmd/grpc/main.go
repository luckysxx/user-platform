package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/luckysxx/user-platform/common/auth"
	"github.com/luckysxx/user-platform/common/config"
	"github.com/luckysxx/user-platform/common/database"
	"github.com/luckysxx/user-platform/common/logger"
	"github.com/luckysxx/user-platform/db"
	pb "github.com/luckysxx/user-platform/proto/user"
	"github.com/luckysxx/user-platform/repository"
	"github.com/luckysxx/user-platform/service"
	usergrpc "github.com/luckysxx/user-platform/transport/grpc"

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
	conn := database.InitPostgres(cfg.Database, logg)
	queries := db.New(conn)
	userRepo := repository.NewUserRepository(queries)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)
	userSvc := service.NewUserService(userRepo, jwtManager, logg)

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, usergrpc.NewUserServer(userSvc, logg))
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
