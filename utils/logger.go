package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	topLeft     = "┌"
	topRight    = "┐"
	bottomLeft  = "└"
	bottomRight = "┘"
	horizontal  = "─"
	vertical    = "│"
)

const (
	Reset       = "\033[0m"
	Bold        = "\033[1m"
	Black       = "\033[30m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BgBlack     = "\033[40m"
	BgRed       = "\033[41m"
	BgGreen     = "\033[42m"
	BgYellow    = "\033[43m"
	BgBlue      = "\033[44m"
	BgMagenta   = "\033[45m"
	BgCyan      = "\033[46m"
	BgWhite     = "\033[47m"
	BrightBlack = "\033[90m"
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
	ServiceName string
	UseColors   bool
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

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		config.UseColors = false
	}

	return AppLogger{
		config: config,
	}
}

func DefaultAppLogger(env string, logLevel string, serviceName string) AppLogger {
	config := LogConfig{
		Environment: env,
		Output:      os.Stdout,
		ErrorOutput: os.Stderr,
		ServiceName: serviceName,
		UseColors:   true,
	}
	if logLevel == "debug" {
		config.Level = DebugLevel
	} else {
		config.Level = InfoLevel
	}
	return NewAppLogger(config)
}

func NewServiceLogger(serviceName string) AppLogger {
	env := os.Getenv("ENV")
	logLevel := os.Getenv("LOG_LEVEL")
	return DefaultAppLogger(env, logLevel, serviceName)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func (l *AppLogger) simpleLog(level LogLevel, prefix, levelStr, color, format string, v ...interface{}) {
	if level < l.config.Level {
		return
	}

	timestamp := time.Now().Format("2006:01:02 15:04:05.000")
	message := fmt.Sprintf(format, v...)

	serviceName := l.config.ServiceName
	if len(serviceName) > 15 {
		serviceName = serviceName[:15]
	}

	logLine := fmt.Sprintf("%s | %-15s | %s | %s", timestamp, serviceName, levelStr, message)

	if level >= ErrorLevel {
		fmt.Fprintln(l.config.ErrorOutput, logLine)
	} else {
		fmt.Fprintln(l.config.Output, logLine)
	}
}

func (l *AppLogger) log(level LogLevel, prefix, levelStr, color, format string, v ...interface{}) {
	if level < l.config.Level {
		return
	}
	if l.config.Environment == "production" || l.config.Environment == "staging" {
		l.simpleLog(level, prefix, levelStr, color, format, v...)
		return
	}

	timestamp := time.Now().Format("2006:01:02 15:04:05.000")
	message := fmt.Sprintf(format, v...)

	termWidth := getTerminalWidth()

	contentWidth := termWidth - 2

	header := fmt.Sprintf(" %s %s %-15s %s %s %s%s ", timestamp, vertical, l.config.ServiceName, vertical, prefix, levelStr, vertical)
	headerWidth := len(header)

	messageLines := strings.Split(message, "\n")
	var formattedLines []string

	firstLineMax := contentWidth - headerWidth
	if len(messageLines) > 0 && len(messageLines[0]) > 0 {
		if len(messageLines[0]) <= firstLineMax {
			formattedLines = append(formattedLines, header+messageLines[0])
		} else {
			formattedLines = append(formattedLines, header+messageLines[0][:firstLineMax])
			remaining := messageLines[0][firstLineMax:]

			for len(remaining) > 0 {
				lineWidth := contentWidth - headerWidth
				if len(remaining) <= lineWidth {
					formattedLines = append(formattedLines, strings.Repeat(" ", headerWidth)+remaining)
					remaining = ""
				} else {
					formattedLines = append(formattedLines, strings.Repeat(" ", headerWidth)+remaining[:lineWidth])
					remaining = remaining[lineWidth:]
				}
			}
		}
	} else {
		formattedLines = append(formattedLines, header)
	}

	for i := 1; i < len(messageLines); i++ {
		line := messageLines[i]
		lineWidth := contentWidth - headerWidth

		for len(line) > 0 {
			if len(line) <= lineWidth {
				formattedLines = append(formattedLines, strings.Repeat(" ", headerWidth)+line)
				line = ""
			} else {
				formattedLines = append(formattedLines, strings.Repeat(" ", headerWidth)+line[:lineWidth])
				line = line[lineWidth:]
			}
		}
	}

	var output strings.Builder

	if l.config.UseColors && color != "" {
		output.WriteString(color)
	}

	output.WriteString(topLeft + strings.Repeat(horizontal, contentWidth) + topRight + "\n")

	for _, line := range formattedLines {
		padding := contentWidth - len(line)
		if padding < 0 {
			line = line[:contentWidth]
			padding = 0
		}
		output.WriteString(vertical + line + strings.Repeat(" ", padding) + "\n")
	}

	output.WriteString(bottomLeft + strings.Repeat(horizontal, contentWidth) + bottomRight)

	if l.config.UseColors && color != "" {
		output.WriteString(Reset)
	}

	output.WriteString("\n")

	if level >= ErrorLevel {
		fmt.Fprint(l.config.ErrorOutput, output.String())
	} else {
		fmt.Fprint(l.config.Output, output.String())
	}
}

func (l *AppLogger) Debug(format string, v ...interface{}) {
	if l.config.Level > DebugLevel {
		return
	}
	l.log(DebugLevel, "🐛", "DEBUG    ", Cyan, format, v...)
}

func (l *AppLogger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, "ℹ️ ", "INFO     ", Blue, format, v...)
}

func (l *AppLogger) Success(format string, v ...interface{}) {
	if l.config.Level > DebugLevel {
		return
	}
	l.log(InfoLevel, "✅", "SUCCESS  ", Green, format, v...)
}

func (l *AppLogger) Warn(format string, v ...interface{}) {
	if l.config.Level > DebugLevel {
		return
	}
	l.log(WarnLevel, "⚠️ ", "WARN     ", Yellow, format, v...)
}

func (l *AppLogger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, "❌", "ERROR    ", Red, format, v...)
}

func (l *AppLogger) Fatal(format string, v ...interface{}) {
	l.log(ErrorLevel, "❌", "FATAL    ", Red, format, v...)
	os.Exit(1)
}

func (l *AppLogger) SetLevel(level LogLevel) {
	l.config.Level = level
}

func (l *AppLogger) SetUseColors(useColors bool) {
	l.config.UseColors = useColors
}
