package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ILogger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	Sync() error
}

var (
	instance ILogger
	once     sync.Once
)

type Logger struct {
	zap *zap.SugaredLogger
}

var NewLogger = func(level string, logFile string) (ILogger, error) {
	once.Do(func() {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		var zapLevel zapcore.Level
		if err := zapLevel.Set(level); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid log level: %s\n", level)
			zapLevel = zapcore.InfoLevel
		}

		var writer zapcore.WriteSyncer
		if logFile != "" {
			writer = zapcore.AddSync(&lumberjack.Logger{
				Filename:   logFile,
				MaxSize:    10,   // megabytes
				MaxBackups: 3,    // number of backups
				MaxAge:     28,   // days
				Compress:   true, // compress the backups
			})
		} else {
			writer = zapcore.AddSync(os.Stdout)
		}

		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			writer,
			zapLevel,
		)

		logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		instance = &Logger{
			zap: logger.Sugar(),
		}
	})

	if instance == nil {
		return nil, fmt.Errorf("failed to create logger instance")
	}

	return instance, nil
}

func addLevelPrefix(level, msg string) string {
	return fmt.Sprintf("[%s] %s", strings.ToUpper(level), msg)
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.zap.Debugw(addLevelPrefix("DEBUG", msg), keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.zap.Infow(addLevelPrefix("INFO", msg), keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.zap.Warnw(addLevelPrefix("WARN", msg), keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.zap.Errorw(addLevelPrefix("ERROR", msg), keysAndValues...)
}

func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.zap.Fatalw(addLevelPrefix("FATAL", msg), keysAndValues...)
}

func (l *Logger) Sync() error {
	if l.zap != nil {
		return l.zap.Sync()
	}
	return nil
}