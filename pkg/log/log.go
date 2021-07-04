package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/juan-leon/fetter/pkg/settings"
)

// Console writes to stderr, in human format.  Direct usage of it is intended
// for errors initializing stuff from cli arguments, or reading the config.  In
// non daemon mode, logs will be forked to this logger.
var Console zap.SugaredLogger

// Logger is the general logger.  In daemon mode, the only logger.
var Logger zap.SugaredLogger

func InitConsoleLogger() {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.Lock(os.Stderr),
		zap.DebugLevel,
	)
	Console = *zap.New(core).Sugar()
}

func InitFileLogger(config settings.Logging) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		Console.Errorf("Setting log level to INFO -> %s", err)
		level.SetLevel(zap.InfoLevel)
	}
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.TimeKey = "@timestamp"
	jsonLog, _, err := zap.Open(config.File)
	if err != nil {
		Console.Fatalf("Failed opening log file: %s", err)
	}
	jsonCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.Lock(jsonLog),
		level,
	)
	core := zapcore.NewTee(
		jsonCore,
		Console.Desugar().Core(),
	)
	Logger = *zap.New(core).Sugar()
}

func InitLoggerForTests() {
	cfg := zap.NewDevelopmentConfig()
	logger, _ := cfg.Build()
	Logger = *logger.Sugar()
}
