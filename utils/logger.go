package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type LogConfig struct {
	Environment string
	Level       LogLevel
	Output      io.Writer
	ErrorOutput io.Writer
}

type AppLogger struct {
	config LogConfig
}

func NewAppLogger(config LogConfig) AppLogger {

	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.ErrorOutput == nil {
		config.ErrorOutput = os.Stderr
	}

	if config.Environment == "production" {

		if config.Level < InfoLevel {
			config.Level = InfoLevel
		}
	}

	return AppLogger{
		config: config,
	}
}

func DefaultAppLogger(environment string) AppLogger {
	config := LogConfig{
		Environment: strings.ToLower(environment),
		Output:      os.Stdout,
		ErrorOutput: os.Stderr,
	}

	if config.Environment == "production" {
		config.Level = InfoLevel
	} else {
		config.Level = DebugLevel
	}

	return NewAppLogger(config)
}

func (l *AppLogger) log(level LogLevel, levelStr, format string, v ...interface{}) {

	if level < l.config.Level {
		return
	}

	timestamp := time.Now().Format("2006:01:02 15:04:05")

	message := fmt.Sprintf(format, v...)

	logLine := fmt.Sprintf("%s  |  %-10s  |  %s\n", timestamp, "[ "+levelStr+" ]", message)

	if level >= ErrorLevel {
		fmt.Fprint(l.config.ErrorOutput, " "+logLine)
	} else {
		fmt.Fprint(l.config.Output, " "+logLine)
	}
}

func (l *AppLogger) Debug(format string, v ...interface{}) {
	l.log(DebugLevel, "🐛 DEBUG", format, v...)
}

func (l *AppLogger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, "ℹ️  INFO ", format, v...)
}

func (l *AppLogger) Warn(format string, v ...interface{}) {
	l.log(WarnLevel, "⚠️  WARN ", format, v...)
}

func (l *AppLogger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, "🔴 ERROR", format, v...)
}

func (l *AppLogger) Fatal(format string, v ...interface{}) {
	l.log(ErrorLevel, "🔴 FATAL", format, v...)
	os.Exit(1)
}

func (l *AppLogger) SetLevel(level LogLevel) {
	l.config.Level = level
}
