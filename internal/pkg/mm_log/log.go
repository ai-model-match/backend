package mm_log

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Define ANSI color codes
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
)

func NewLogger(appMode string) *zap.Logger {
	var logger *zap.Logger
	if appMode != "release" {
		// Custom level encoder with colors
		levelEncoder := func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			var lvl string
			switch l {
			case zapcore.DebugLevel:
				lvl = blue + "DEBUG" + reset
			case zapcore.InfoLevel:
				lvl = green + "INFO" + reset
			case zapcore.WarnLevel:
				lvl = yellow + "WARN" + reset
			case zapcore.ErrorLevel:
				lvl = red + "ERROR" + reset
			case zapcore.DPanicLevel:
				lvl = magenta + "DPANIC" + reset
			case zapcore.PanicLevel:
				lvl = magenta + "PANIC" + reset
			case zapcore.FatalLevel:
				lvl = red + "FATAL" + reset
			default:
				lvl = l.String()
			}
			enc.AppendString(lvl)
		}
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = levelEncoder
		// Customize info colours
		cfg.EncoderConfig.EncodeCaller = zapcore.CallerEncoder(func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			path := strings.Split(caller.TrimmedPath(), ":")[0]
			line := caller.Line
			enc.AppendString(fmt.Sprintf("%s%s%s:%s%d%s",
				blue, path, reset, yellow, line, reset,
			))
		})
		// Customize date format and colours
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			formatted := t.Format("2006-01-02 15:04:05.000")
			enc.AppendString(cyan + formatted + reset)
		})
		// Create new logger
		logger = zap.Must(cfg.Build())
	} else {
		logger = zap.Must(zap.NewDevelopment())
	}
	defer logger.Sync()
	return logger
}

func LogPanicError(r any, serviceName string, message string) {
	buff := make([]byte, 1<<16)
	n := runtime.Stack(buff, false)
	zap.L().Error(message, zap.String("service", serviceName), zap.Any("panic", r), zap.ByteString("stack", buff[:n]))
}
