package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/natefinch/lumberjack"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

var fieldsLevelMap = map[string]logrus.Level{
	"trace": logrus.TraceLevel,
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"fatal": logrus.FatalLevel,
	"panic": logrus.PanicLevel,
}

func NewLogger(cfg *config.Config) *logrus.Logger {
	Log = logrus.New()

	level, exist := fieldsLevelMap[cfg.Logger.Level]
	if !exist {
		fmt.Printf("unknown logger level: %s\n", cfg.Logger.Level)
		level = logrus.DebugLevel
	}

	if cfg.Logger.Encoding != "text" && cfg.Logger.Encoding != "json" {
		fmt.Printf("unknown logger encoding: %s\n", cfg.Logger.Encoding)
		cfg.Logger.Encoding = "text"
	}

	var formatter logrus.Formatter
	if cfg.Logger.Encoding == "text" {
		formatter = &logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   false,

			CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
				file = fmt.Sprintf("%s:%d", f.File, f.Line)
				return function, filepath.Base(file)
			},
		}
	} else {
		formatter = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			PrettyPrint:     true,
			CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
				file = fmt.Sprintf("%s:%d", f.File, f.Line)
				return function, filepath.Base(file)
			},
		}
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   "./logs/application.log",
		MaxSize:    10,
		MaxBackups: 10,
		MaxAge:     28,
		Compress:   true,
	}

	mw := io.MultiWriter(os.Stdout, lumberjackLogger)
	Log.SetLevel(level)
	Log.SetFormatter(formatter)
	Log.SetReportCaller(cfg.Logger.Caller)
	Log.SetOutput(mw)

	// Configure global logrus standard logger
	// logrus.SetLevel(level)
	// logrus.SetFormatter(formatter)
	// logrus.SetReportCaller(cfg.Logger.Caller)
	// logrus.SetOutput(os.Stdout)
	// logrus.SetOutput(lumberjackLogger)

	return Log
}
