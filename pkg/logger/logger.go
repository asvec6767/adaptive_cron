package logger

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"adaptive-cron/internal/common/models"
)

var (
	log          = logrus.New()
	clickHouse   ClickHouseWriter
	mu           sync.RWMutex
	isClickHouse bool
	serviceName  string
)

// ClickHouseWriter интерфейс для записи в ClickHouse
type ClickHouseWriter interface {
	WriteLog(ctx context.Context, log *models.ClickHouseLog) error
	WriteBatch(ctx context.Context, logs []models.ClickHouseLog) error
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level            string
	Format           string
	Service          string
	ClickHouseDSN    string
	EnableClickHouse bool
}

// LogEntry структура лога
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	TaskID     string                 `json:"task_id,omitempty"`
	TaskName   string                 `json:"task_name,omitempty"`
	JobID      string                 `json:"job_id,omitempty"`
	DurationMs int64                  `json:"duration_ms,omitempty"`
	ExitCode   int                    `json:"exit_code,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
}

func Init(config LoggerConfig) {
	// Уровень логирования
	lvl, err := logrus.ParseLevel(config.Level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)

	// Формат вывода
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	log.SetOutput(os.Stdout)
	serviceName = config.Service
	isClickHouse = config.EnableClickHouse

	// Инициализация ClickHouse если включен
	if isClickHouse && config.ClickHouseDSN != "" {
		// Инициализация будет через отдельную функцию
	}
}

// SetClickHouseWriter устанавливает writer для ClickHouse
func SetClickHouseWriter(writer ClickHouseWriter) {
	mu.Lock()
	defer mu.Unlock()
	clickHouse = writer
}

// LogEvent структура события для логирования
type LogEvent struct {
	Message    string
	Level      string
	TaskID     string
	TaskName   string
	JobID      string
	DurationMs int64
	ExitCode   int
	Error      string
	CPUPercent float64
	MemoryMB   int64
	LoadScore  float64
	Fields     map[string]interface{}
}

// LogWithEvent логирует событие с дополнительными данными
func LogWithEvent(ctx context.Context, event LogEvent) {
	// Логируем в stdout
	entry := log.WithFields(logrus.Fields{})

	// Добавляем поля
	if event.TaskID != "" {
		entry = entry.WithField("task_id", event.TaskID)
	}
	if event.TaskName != "" {
		entry = entry.WithField("task_name", event.TaskName)
	}
	if event.JobID != "" {
		entry = entry.WithField("job_id", event.JobID)
	}
	if event.DurationMs > 0 {
		entry = entry.WithField("duration_ms", event.DurationMs)
	}
	if event.ExitCode != 0 {
		entry = entry.WithField("exit_code", event.ExitCode)
	}
	if event.Error != "" {
		entry = entry.WithField("error", event.Error)
	}
	if event.Fields != nil {
		for k, v := range event.Fields {
			entry = entry.WithField(k, v)
		}
	}

	// Логируем в зависимости от уровня
	switch strings.ToLower(event.Level) {
	case "debug":
		entry.Debug(event.Message)
	case "info":
		entry.Info(event.Message)
	case "warn", "warning":
		entry.Warn(event.Message)
	case "error":
		entry.Error(event.Message)
	case "fatal":
		entry.Fatal(event.Message)
	default:
		entry.Info(event.Message)
	}

	// Отправляем в ClickHouse
	if isClickHouse && clickHouse != nil {
		go func() {
			chLog := &models.ClickHouseLog{
				Timestamp:  time.Now(),
				Level:      event.Level,
				Service:    serviceName,
				TaskID:     event.TaskID,
				TaskName:   event.TaskName,
				JobID:      event.JobID,
				Message:    event.Message,
				Error:      event.Error,
				DurationMs: event.DurationMs,
				ExitCode:   event.ExitCode,
				CPUPercent: event.CPUPercent,
				MemoryMB:   event.MemoryMB,
				LoadScore:  event.LoadScore,
				RequestID:  GetRequestID(ctx),
				Tags:       getTags(event),
			}

			// Добавляем метаданные
			if event.Fields != nil {
				if metadata, err := json.Marshal(event.Fields); err == nil {
					chLog.Metadata = string(metadata)
				}
			}

			if err := clickHouse.WriteLog(context.Background(), chLog); err != nil {
				log.WithError(err).Error("Failed to write log to ClickHouse")
			}
		}()
	}
}

// getTags извлекает теги из события
func getTags(event LogEvent) []string {
	tags := []string{}
	if event.Level != "" {
		tags = append(tags, "level:"+event.Level)
	}
	if event.TaskName != "" {
		tags = append(tags, "task:"+event.TaskName)
	}
	if event.ExitCode == 0 && event.Level != "error" {
		tags = append(tags, "success")
	}
	if event.ExitCode != 0 {
		tags = append(tags, "failure")
	}
	return tags
}

// GetRequestID получает RequestID из контекста
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// GenerateRequestID генерирует новый RequestID
func GenerateRequestID() string {
	return uuid.New().String()
}

// WithRequestID добавляет RequestID в контекст
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// Convenience functions
func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

func WithFields(fields map[string]interface{}) *logrus.Entry {
	return log.WithFields(fields)
}

// Debugf форматированный debug лог
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Infof форматированный info лог
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warnf форматированный warn лог
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Errorf форматированный error лог
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatalf форматированный fatal лог
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// GetCallerInfo получает информацию о вызывающем
func GetCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	return file, line
}
