package database

import (
	"context"

	_ "github.com/lib/pq"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/platform/config"

	"go.uber.org/zap"
)

// InitEntClient 初始化 Ent 客户端并验证数据库连接配置。
func InitEntClient(cfg config.DatabaseConfig, log *zap.Logger) *ent.Client {
	client, err := ent.Open(cfg.Driver, cfg.Source)
	if err != nil {
		log.Fatal("无法初始化 Ent 客户端", zap.Error(err))
		return nil
	}

	if cfg.AutoMigrate {
		if err := client.Schema.Create(context.Background()); err != nil {
			log.Fatal("自动执行 Ent schema migration 失败", zap.Error(err))
			return nil
		}
		log.Info("已执行 Ent schema migration", zap.Bool("auto_migrate", true))
	} else {
		log.Info("跳过 Ent schema migration", zap.Bool("auto_migrate", false))
	}

	log.Info("成功初始化 Ent 客户端")
	return client
}
