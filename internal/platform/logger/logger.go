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
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}

	if env == "dev" || env == "development" {
		level = zapcore.DebugLevel
		// 开发环境加点颜色高亮，方便人眼阅读
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 选择编码器：容器环境用 JSON，本地用 Console
	isContainer := env == "production" || env == "prod" || env == "container"
	var encoder zapcore.Encoder
	if isContainer {
		encoder = zapcore.NewJSONEncoder(config)
	} else {
		encoder = zapcore.NewConsoleEncoder(config)
	}

	// 判断日志输出方式
	var writeSyncer zapcore.WriteSyncer
	logFile := os.Getenv("LOG_FILE")

	if isContainer && logFile == "" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		if logFile == "" {
			logFile = "app.log"
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else {
			writeSyncer = zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(file),
			)
		}
	}

	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		level,
	)

	// AddCaller添加调用者信息。去掉普通 Error 的 Stacktrace 避免输出一堆 github.com 的堆栈信息
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.DPanicLevel), // 仅在 Panic 时打印堆栈
	)

	logger = logger.With(zap.String("service", serviceName))

	return logger
}
