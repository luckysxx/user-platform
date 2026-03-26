package database

import (
	"context"

	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/migrate"
	"github.com/luckysxx/user-platform/internal/platform/config"

	"go.uber.org/zap"
)

// InitEntClient 初始化 Ent 客户端并验证数据库连接配置。
func InitEntClient(cfg config.DatabaseConfig, log *zap.Logger) *ent.Client {
	db, err := otelsql.Open(cfg.Driver, cfg.Source,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
	)
	if err != nil {
		log.Fatal("无法初始化带 OTel 追踪的数据库连接", zap.Error(err))
		return nil
	}

	drv := sql.OpenDB(cfg.Driver, db)
	client := ent.NewClient(ent.Driver(drv))

	if cfg.AutoMigrate {
		if err := client.Schema.Create(
			context.Background(),
			migrate.WithDropIndex(true),
			migrate.WithDropColumn(true),
		); err != nil {
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

