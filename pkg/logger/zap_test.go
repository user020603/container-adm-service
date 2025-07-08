package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func resetOnce() {
	once = sync.Once{}
	instance = nil
}

func TestNewLogger_Singleton(t *testing.T) {
	resetOnce()
	defer resetOnce()

	logFile := "singleton_test.log"
	defer os.Remove(logFile)

	logger1, err := NewLogger("info", logFile)
	require.NoError(t, err)
	require.NotNil(t, logger1)

	logger2, err := NewLogger("debug", "another_file.log")
	require.NoError(t, err)
	require.NotNil(t, logger2)

	assert.Same(t, logger1, logger2, "NewLogger should return the same instance for multiple calls")

	logger1.Info("This is a test log message")
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "This is a test log message", "Log message should be written to the log file")
}

func TestNewLogger_FileLogging(t *testing.T) {
	resetOnce()
	defer resetOnce()

	logFile := "test_file_logging.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("debug", logFile)
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("This is a file log message", "key", "value")
	err = logger.Sync()
	require.NoError(t, err)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(content), "[INFO] This is a file log message", "Log message should be written to the log file")
	assert.Contains(t, string(content), `"key": "value"`)
}

func TestNewLogger_ConsoleLogging(t *testing.T) {
	resetOnce()
	defer resetOnce()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, err := NewLogger("info", "")
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("Logging to console")

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	assert.Contains(t, output, "[INFO] Logging to console")
}

func TestNewLogger_InvalidLogLevel(t *testing.T) {
	resetOnce()
	defer resetOnce()

	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	logger, err := NewLogger("invalidLevel", "")
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Debug("This should not be logged")
	logger.Info("This should be logged")

	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout

	var errBuf bytes.Buffer
	io.Copy(&errBuf, rErr)
	assert.Contains(t, errBuf.String(), "Invalid log level: invalidLevel")

	var outBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	output := outBuf.String()
	assert.NotContains(t, output, "This should not be logged")
	assert.Contains(t, output, "[INFO] This should be logged")
}

func TestLoggingLevels(t *testing.T) {
	resetOnce()
	defer resetOnce()

	logFile := "test_levels.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("debug", logFile)
	require.NoError(t, err)

	logger.Debug("debug message", "id", 1)
	logger.Info("info message", "id", 2)
	logger.Warn("warn message", "id", 3)
	logger.Error("error message", "id", 4)

	err = logger.Sync()
	require.NoError(t, err)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	logContent := string(content)

	assert.Contains(t, logContent, "[DEBUG] debug message")
	assert.Contains(t, logContent, `"id": 1`)
	assert.Contains(t, logContent, "[INFO] info message")
	assert.Contains(t, logContent, `"id": 2`)
	assert.Contains(t, logContent, "[WARN] warn message")
	assert.Contains(t, logContent, `"id": 3`)
	assert.Contains(t, logContent, "[ERROR] error message")
	assert.Contains(t, logContent, `"id": 4`)
}

func TestFatal(t *testing.T) {
	assert.Panics(t, func() {
		resetOnce()
		defer resetOnce()

		encoderConfig := zap.NewProductionEncoderConfig()
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(io.Discard), // write to nowhere
			zapcore.DebugLevel,
		)

		hookedCore := zapcore.RegisterHooks(core, func(e zapcore.Entry) error {
			if e.Level == zapcore.FatalLevel {
				panic(fmt.Sprintf("fatal hook: %s", e.Message))
			}
			return nil
		})

		zapLogger := zap.New(hookedCore, zap.AddCaller(), zap.AddCallerSkip(1))
		logger := &Logger{zap: zapLogger.Sugar()}

		logger.Fatal("this is a fatal error")
	}, "Expected Fatal to cause a panic via the test hook")
}

func TestSync_NilLogger(t *testing.T) {
	// This tests the guard clause `if l.zap != nil`
	logger := &Logger{zap: nil}
	err := logger.Sync()
	assert.NoError(t, err, "Sync on a logger with a nil zap field should not return an error")
}

func TestAddLevelPrefix(t *testing.T) {
	msg := "test message"
	level := "level"
	expected := "[LEVEL] test message"
	result := addLevelPrefix(level, msg)
	assert.Equal(t, expected, result)
	assert.Equal(t, "[DEBUG] another", addLevelPrefix("debug", "another"))
}
