package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv      string            `mapstructure:"app_env"`
	Server      ServerConfig      `mapstructure:"server"`
	GRPCServer  GRPCServerConfig  `mapstructure:"grpc_server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Kafka       KafkaConfig       `mapstructure:"kafka"`
	IDGenerator IDGeneratorConfig `mapstructure:"id_generator"`
	OTel        OTelConfig        `mapstructure:"otel"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
}

// MetricsConfig Prometheus 指标端口配置
type MetricsConfig struct {
	Port string `mapstructure:"port"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers             string `mapstructure:"brokers"`
	TopicUserRegistered string `mapstructure:"topic_user_registered"`
}

type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`
	Source      string `mapstructure:"source"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type GRPCServerConfig struct {
	Port string `mapstructure:"port"`
}

type IDGeneratorConfig struct {
	Addr string `mapstructure:"addr"`
}

type OTelConfig struct {
	JaegerEndpoint string `mapstructure:"jaeger_endpoint"`
	ServiceName    string `mapstructure:"service_name"`
}

// LoadConfig 从 Viper 加载配置
func LoadConfig() *Config {
	// 在优先加载环境变量之前，尝试从根目录读取 .env 文件（如果在生产环境一般不用此文件，也不会报错）
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs") // 支持项目根目录启动
	viper.AddConfigPath(".")

	// 允许环境变量覆盖配置 (比如 export DATABASE_SOURCE=xyz)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: No config.yaml found, relying entirely on ENV variables: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	return &cfg
}
