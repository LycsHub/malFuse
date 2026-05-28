package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Level  string
	Format string
	Output string
}

var logrusLogger *logrus.Logger

func Init(cfg Config) {
	logrusLogger = logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrusLogger.SetLevel(level)

	if cfg.Format == "json" {
		logrusLogger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrusLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			DisableColors:   true,
			DisableQuote:    true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	if cfg.Output == "stdout" || cfg.Output == "" {
		logrusLogger.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logrusLogger.SetOutput(os.Stdout)
			logrusLogger.Warnf("failed to open log file %s: %v, falling back to stdout", cfg.Output, err)
		} else {
			logrusLogger.SetOutput(io.MultiWriter(os.Stdout, f))
		}
	}
}

func withFields(keysAndValues ...interface{}) *logrus.Entry {
	fields := logrus.Fields{}
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields[key] = keysAndValues[i+1]
	}
	return logrusLogger.WithFields(fields)
}

func Debug(msg string, keysAndValues ...interface{}) {
	if logrusLogger == nil {
		return
	}
	withFields(keysAndValues...).Debug(msg)
}

func Info(msg string, keysAndValues ...interface{}) {
	if logrusLogger == nil {
		return
	}
	withFields(keysAndValues...).Info(msg)
}

func Warn(msg string, keysAndValues ...interface{}) {
	if logrusLogger == nil {
		return
	}
	withFields(keysAndValues...).Warn(msg)
}

func Error(msg string, keysAndValues ...interface{}) {
	if logrusLogger == nil {
		return
	}
	withFields(keysAndValues...).Error(msg)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	if logrusLogger == nil {
		logrus.New().Fatal(msg)
		return
	}
	withFields(keysAndValues...).Fatal(msg)
}

func Reset() {
	logrusLogger = nil
}
