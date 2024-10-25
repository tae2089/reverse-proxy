package log

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	encoderCfg := getEncoderConfig()
	logger, _ = getConfig(encoderCfg).Build()
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Errorf(err error, fields ...zap.Field) {
	logger.Error(err.Error(), fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func getEncoderConfig() zapcore.EncoderConfig {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	// encoderCfg.MessageKey = zapcore.OmitKey
	encoderCfg.LevelKey = zapcore.OmitKey
	return encoderCfg
}

func getConfig(encoderConfig zapcore.EncoderConfig) zap.Config {
	return zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderConfig,
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}
}

func GetLoggingFieldsByRequest(r *http.Request, statusCode int, duration time.Duration) []zap.Field {
	var query string = ""
	if r.URL.RawQuery != "" {
		query = "?" + r.URL.RawQuery
	}
	loggingDefault := []zap.Field{
		zap.Int("Status Code", statusCode),
		zap.String("Remote-Host", r.RemoteAddr),
		zap.String("Method", r.Method),
		zap.String("Path", r.URL.Path+query),
		zap.Duration("Response Time", duration),
	}

	return loggingDefault
}
