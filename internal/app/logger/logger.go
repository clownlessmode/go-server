package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
)

var mu sync.Mutex

type Logger struct {
	component string
}

func New(component string) *Logger {
	return &Logger{component: component}
}

func (l *Logger) Infof(format string, args ...any) {
	l.write("INFO", colorCyan, format, args...)
}

func (l *Logger) Successf(format string, args ...any) {
	l.write("OK", colorGreen, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.write("WARN", colorYellow, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.write("ERROR", colorRed, format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.Errorf(format, args...)
	os.Exit(1)
}

func (l *Logger) write(level string, color string, format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()

	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	component := strings.ToUpper(l.component)

	if colorsEnabled() {
		fmt.Fprintf(
			os.Stdout,
			"%s%s%s %s%-5s%s %s%-10s%s %s",
			colorDim,
			timestamp,
			colorReset,
			color,
			level,
			colorReset,
			colorBlue,
			component,
			colorReset,
			message,
		)
		fmt.Fprintln(os.Stdout)
		return
	}

	fmt.Fprintf(os.Stdout, "%s %-5s %-10s %s\n", timestamp, level, component, message)
}

func colorsEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}
