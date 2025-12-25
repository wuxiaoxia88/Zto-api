package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	file       *os.File
	logger     *log.Logger
	mu         sync.Mutex
	logDir     string
	currentDay string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init 初始化日志系统
func Init(logDir string) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(logDir, INFO)
	})
	return err
}

// NewLogger 创建新的日志记录器
func NewLogger(logDir string, level LogLevel) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	l := &Logger{
		level:  level,
		logDir: logDir,
	}

	if err := l.rotateIfNeeded(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Logger) rotateIfNeeded() error {
	today := time.Now().Format("2006-01-02")
	if l.currentDay == today && l.file != nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	logPath := filepath.Join(l.logDir, fmt.Sprintf("service_%s.log", today))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.currentDay = today

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.logger = log.New(multiWriter, "", 0)

	return nil
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.rotateIfNeeded()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s] ", timestamp, levelNames[level])
	message := fmt.Sprintf(format, args...)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(prefix + message)
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, format, args...)
	}
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, format, args...)
	}
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, format, args...)
	}
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, args...)
	}
}

// Token Token 相关日志
func Token(format string, args ...interface{}) {
	Info("[TOKEN] "+format, args...)
}

// API API 相关日志
func API(format string, args ...interface{}) {
	Info("[API] "+format, args...)
}

// Close 关闭日志器
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}
