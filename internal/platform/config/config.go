package config

import (
	"github.com/luckysxx/common/conf"
	commonOtel "github.com/luckysxx/common/otel"
	"github.com/luckysxx/common/postgres"
	commonRedis "github.com/luckysxx/common/redis"
)

type Config struct {
	AppEnv      string                `mapstructure:"app_env"`
	Server      conf.ServerConfig     `mapstructure:"server"`
	GRPCServer  GRPCServerConfig      `mapstructure:"grpc_server"`
	Database    postgres.Config       `mapstructure:"database"`
	Redis       commonRedis.Config    `mapstructure:"redis"`
	JWT         JWTConfig             `mapstructure:"jwt"`
	Kafka       KafkaConfig           `mapstructure:"kafka"`
	IDGenerator conf.IDGeneratorConfig `mapstructure:"id_generator"`
	OTel        commonOtel.Config     `mapstructure:"otel"`
	Metrics     MetricsConfig         `mapstructure:"metrics"`
}

// === 以下为服务专有配置，不提取到 common ===

type MetricsConfig struct {
	Port string `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers             string `mapstructure:"brokers"`
	TopicUserRegistered string `mapstructure:"topic_user_registered"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type GRPCServerConfig struct {
	Port string `mapstructure:"port"`
}

// LoadConfig 从 Viper 加载配置（底层由 common/conf.Load 统一处理）
func LoadConfig() *Config {
	var cfg Config
	conf.Load(&cfg)
	return &cfg
}
