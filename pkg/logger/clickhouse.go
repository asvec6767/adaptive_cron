package logger

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"adaptive-cron/internal/common/models"
)

type ClickHouseWriterImpl struct {
	conn      driver.Conn
	batchSize int
	batch     []models.ClickHouseLog
	mu        sync.Mutex
	stopChan  chan struct{}
}

var (
	clickHouseWriter *ClickHouseWriterImpl
	once             sync.Once
)

// NewClickHouseWriter создает новый ClickHouse writer
func NewClickHouseWriter(dsn string, batchSize int) (*ClickHouseWriterImpl, error) {
	var err error
	once.Do(func() {
		clickHouseWriter, err = createClickHouseWriter(dsn, batchSize)
	})
	return clickHouseWriter, err
}

func createClickHouseWriter(dsn string, batchSize int) (*ClickHouseWriterImpl, error) {
	// Подключение к ClickHouse
	conn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	writer := &ClickHouseWriterImpl{
		conn:      conn.(driver.Conn),
		batchSize: batchSize,
		batch:     make([]models.ClickHouseLog, 0, batchSize),
		stopChan:  make(chan struct{}),
	}

	// Создание таблицы
	if err := writer.createTable(); err != nil {
		return nil, err
	}

	// Запуск фонового процесса для периодической записи
	go writer.backgroundFlush()

	Info("ClickHouse writer initialized successfully")
	return writer, nil
}

func (w *ClickHouseWriterImpl) createTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS logs (
        timestamp DateTime,
        level String,
        service String,
        task_id String,
        task_name String,
        job_id String,
        message String,
        error String,
        duration_ms Int64,
        exit_code Int16,
        cpu_percent Float64,
        memory_mb Int64,
        load_score Float64,
        ip_address String,
        user_agent String,
        request_id String,
        tags Array(String),
        metadata String
    ) ENGINE = MergeTree()
    ORDER BY (timestamp, service, level)
    SETTINGS index_granularity = 8192
    `

	// Создаем таблицу если не существует
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return w.conn.Exec(ctx, query)
}

// WriteLog записывает один лог
func (w *ClickHouseWriterImpl) WriteLog(ctx context.Context, log *models.ClickHouseLog) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.batch = append(w.batch, *log)

	if len(w.batch) >= w.batchSize {
		return w.flush(ctx)
	}

	return nil
}

// WriteBatch записывает батч логов
func (w *ClickHouseWriterImpl) WriteBatch(ctx context.Context, logs []models.ClickHouseLog) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.batch = append(w.batch, logs...)

	if len(w.batch) >= w.batchSize {
		return w.flush(ctx)
	}

	return nil
}

// flush отправляет батч в ClickHouse
func (w *ClickHouseWriterImpl) flush(ctx context.Context) error {
	if len(w.batch) == 0 {
		return nil
	}

	batch, err := w.conn.PrepareBatch(ctx, `
        INSERT INTO logs (
            timestamp, level, service, task_id, task_name, job_id,
            message, error, duration_ms, exit_code, cpu_percent,
            memory_mb, load_score, ip_address, user_agent,
            request_id, tags, metadata
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, log := range w.batch {
		if err := batch.Append(
			log.Timestamp,
			log.Level,
			log.Service,
			log.TaskID,
			log.TaskName,
			log.JobID,
			log.Message,
			log.Error,
			log.DurationMs,
			log.ExitCode,
			log.CPUPercent,
			log.MemoryMB,
			log.LoadScore,
			log.IPAddress,
			log.UserAgent,
			log.RequestID,
			log.Tags,
			log.Metadata,
		); err != nil {
			return fmt.Errorf("failed to append log: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	w.batch = w.batch[:0]
	return nil
}

// backgroundFlush периодически сбрасывает батч
func (w *ClickHouseWriterImpl) backgroundFlush() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			w.mu.Lock()
			err := w.flush(ctx)
			w.mu.Unlock()
			cancel()

			if err != nil {
				Errorf("Failed to flush logs to ClickHouse: %v", err)
			}
		case <-w.stopChan:
			// Финальный сброс перед остановкой
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			w.mu.Lock()
			err := w.flush(ctx)
			w.mu.Unlock()
			cancel()

			if err != nil {
				Errorf("Failed to flush logs on stop: %v", err)
			}
			return
		}
	}
}

// Close закрывает writer
func (w *ClickHouseWriterImpl) Close() error {
	close(w.stopChan)
	return w.conn.Close()
}
