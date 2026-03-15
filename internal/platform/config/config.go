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
}

type ServerConfig struct {
	Port string
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

	// 数据库配置：使用完整的连接字符串
	// 格式：postgres://user:password@host:port/dbname?sslmode=disable
	// 支持其他数据库：mysql://user:password@host:port/dbname
	dbSource := getEnv("DB_SOURCE",
		"postgres://luckys:123456@localhost:5432/gopher_paste?sslmode=disable")
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
			Secret: getEnv("JWT_SECRET", "gopherpaste_secret_key"),
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
