package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 创建一个新的 Logger 实例
// serviceName 用于标识日志来源的服务名称
func NewLogger(serviceName string) *zap.Logger {
	config := zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		CallerKey:     "caller",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",

		EncodeTime:   zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	// 开发环境使用Debug级别，生产环境使用Info级别
	level := zapcore.InfoLevel
	env := os.Getenv("ENV")
	if env == "dev" || env == "development" {
		level = zapcore.DebugLevel
	}

	// 判断日志输出方式
	// 1. 容器环境 (ENV=production/prod/container)：只输出到 stdout
	// 2. 本地开发环境：同时输出到 stdout 和文件
	// 3. 可通过 LOG_FILE 环境变量自定义文件路径
	var writeSyncer zapcore.WriteSyncer

	isContainer := env == "production" || env == "prod" || env == "container"
	logFile := os.Getenv("LOG_FILE")

	if isContainer && logFile == "" {
		// 容器环境且未指定文件：只输出到 stdout
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		// 本地开发环境或指定了 LOG_FILE：同时输出到 stdout 和文件
		if logFile == "" {
			logFile = "app.log" // 本地开发默认文件
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// 文件打开失败，降级到只输出 stdout
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else {
			writeSyncer = zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(file),
			)
		}
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		writeSyncer,
		level,
	)

	// AddCaller添加调用者信息，AddStacktrace在Error级别添加堆栈
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	logger = logger.With(zap.String("service", serviceName))

	return logger
}
