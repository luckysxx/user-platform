package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Port string
}

type KafkaConfig struct {
	Brokers             string
	TopicUserRegistered string
}

type DatabaseConfig struct {
	Driver      string
	Source      string
	AutoMigrate bool
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
}

// LoadConfig 从环境变量加载配置（后期可改为配置文件）
func LoadConfig() *Config {
	// 尝试加载 .env 文件（如果存在）
	_ = godotenv.Load()

	dbSource := getEnv("DB_SOURCE",
		"postgres://luckys:123456@localhost:5432/user_platform?sslmode=disable")
	appEnv := getEnv("APP_ENV", "development")
	autoMigrateDefault := appEnv != "production"

	return &Config{
		AppEnv: appEnv,
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Driver:      getEnv("DB_DRIVER", "postgres"),
			Source:      dbSource,
			AutoMigrate: getEnvAsBool("DB_AUTO_MIGRATE", autoMigrateDefault),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", "123456"),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "user_platform_secret_key"),
		},
		Kafka: KafkaConfig{
			Brokers:             getEnv("KAFKA_BROKERS", "localhost:9092"),
			TopicUserRegistered: getEnv("KAFKA_TOPIC_USER_REGISTERED", "user_registered"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
