package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv   string         `mapstructure:"app_env"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
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

// LoadConfig 从 Viper 加载配置
func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs") // 支持项目根目录启动
	viper.AddConfigPath(".")         
	
	// 允许环境变量覆盖配置 (比如 export KAFKA_BROKERS=xyz)
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
