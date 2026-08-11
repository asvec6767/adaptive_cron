package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"adaptive-cron/internal/common/models"
	"adaptive-cron/internal/common/restclient"
	"adaptive-cron/pkg/logger"
)

type Scheduler struct {
	db            *gorm.DB
	redis         *redis.Client
	metricClient  *restclient.Client
	workerClient  *restclient.Client
	router        *mux.Router
	server        *http.Server
	checkInterval time.Duration
	stopCh        chan struct{}
}

type DelayedTask struct {
	TaskID     string
	DelayUntil time.Time
	Attempt    int
}

func NewScheduler(db *gorm.DB, redis *redis.Client, metricURL, workerURL string, checkInterval time.Duration) *Scheduler {
	s := &Scheduler{
		db:            db,
		redis:         redis,
		metricClient:  restclient.NewClient(metricURL, 5*time.Second),
		workerClient:  restclient.NewClient(workerURL, 30*time.Second),
		router:        mux.NewRouter(),
		checkInterval: checkInterval,
		stopCh:        make(chan struct{}),
	}
	s.setupRoutes()
	return s
}

func (s *Scheduler) setupRoutes() {
	api := s.router.PathPrefix("/api/v1/scheduler").Subrouter()
	api.HandleFunc("/next", s.getNextTasks).Methods("GET")
	api.HandleFunc("/trigger/{taskId}", s.triggerTask).Methods("POST")
	api.HandleFunc("/queue", s.getDelayedQueue).Methods("GET")
	api.HandleFunc("/queue/{taskId}", s.removeFromQueue).Methods("DELETE")

	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
}

func (s *Scheduler) getNextTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []models.Task
	if err := s.db.Where("is_active = ?", true).Find(&tasks).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}

	nextTasks := make([]map[string]interface{}, 0)
	now := time.Now()

	for _, task := range tasks {
		// Simple cron check (in production use robfig/cron)
		if s.shouldRunNow(now, task.Schedule) {
			nextTasks = append(nextTasks, map[string]interface{}{
				"id":          task.ID,
				"name":        task.Name,
				"command":     task.Command,
				"is_critical": task.IsCritical,
				"schedule":    task.Schedule,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": nextTasks,
		"count": len(nextTasks),
	})
}

func (s *Scheduler) triggerTask(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskId"]
	force := r.URL.Query().Get("force") == "true"

	var task models.Task
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	// Adaptive check
	if !force && !task.IsCritical {
		canRun, delay, reason := s.checkResourceAvailability(&task)
		if !canRun {
			// Add to delayed queue
			delayedTask := DelayedTask{
				TaskID:     task.ID,
				DelayUntil: time.Now().Add(delay),
				Attempt:    1,
			}
			s.addToDelayedQueue(delayedTask)

			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"status": "delayed",
				"reason": reason,
				"delay":  delay.String(),
			})
			return
		}
	}

	// Execute task
	jobID, err := s.executeTask(&task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "executed",
		"job_id": jobID,
	})
}

func (s *Scheduler) getDelayedQueue(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get all tasks from Redis sorted set
	now := time.Now().Unix()
	tasks, err := s.redis.ZRangeByScore(ctx, "delayed_queue", &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", now),
		Max: "+inf",
	}).Result()

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get delayed queue")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func (s *Scheduler) removeFromQueue(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskId"]
	ctx := context.Background()

	// Remove from Redis
	if err := s.redis.ZRem(ctx, "delayed_queue", taskID).Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to remove from queue")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *Scheduler) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database
	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		writeError(w, http.StatusServiceUnavailable, "Database unhealthy")
		return
	}

	// Check Redis
	ctx := context.Background()
	if err := s.redis.Ping(ctx).Err(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "Redis unhealthy")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Scheduler) checkResourceAvailability(task *models.Task) (bool, time.Duration, string) {
	ctx := context.Background()

	// Get current metrics
	var metrics models.SystemMetrics
	if err := s.metricClient.Get(ctx, "/api/v1/metrics/current", &metrics); err != nil {
		logger.WithField("error", err).Warn("Failed to get metrics, allowing execution")
		return true, 0, "metrics_unavailable"
	}

	// Check CPU
	if metrics.CPUPercent > task.MaxCPUPercent {
		delay := 30 * time.Second
		return false, delay, fmt.Sprintf("CPU high: %.2f%% > %.2f%%", metrics.CPUPercent, task.MaxCPUPercent)
	}

	// Check Memory
	memoryUsedMB := int64(metrics.MemoryUsed / 1024 / 1024)
	if memoryUsedMB > task.MaxMemoryMB {
		delay := 30 * time.Second
		return false, delay, fmt.Sprintf("Memory high: %dMB > %dMB", memoryUsedMB, task.MaxMemoryMB)
	}

	// Check Load Score
	loadScore := metrics.GetLoadScore()
	if loadScore > task.MaxLoadScore {
		delay := time.Duration(loadScore*60) * time.Second
		return false, delay, fmt.Sprintf("Load score high: %.2f > %.2f", loadScore, task.MaxLoadScore)
	}

	return true, 0, "sufficient_resources"
}

func (s *Scheduler) executeTask(task *models.Task) (string, error) {
	ctx := context.Background()

	type ExecuteResponse struct {
		JobID string `json:"job_id"`
	}

	req := map[string]interface{}{
		"task_id": task.ID,
		"command": task.Command,
		"timeout": task.TimeoutSeconds,
	}

	var resp ExecuteResponse
	if err := s.workerClient.Post(ctx, "/api/v1/execute", req, &resp); err != nil {
		return "", err
	}

	// Log execution start
	jobLog := &models.JobLog{
		TaskID:    task.ID,
		TaskName:  task.Name,
		StartTime: time.Now(),
	}
	s.db.Create(jobLog)

	logger.WithFields(map[string]interface{}{
		"task_id":   task.ID,
		"task_name": task.Name,
		"job_id":    resp.JobID,
	}).Info("Task execution started")

	return resp.JobID, nil
}

func (s *Scheduler) shouldRunNow(now time.Time, schedule string) bool {
	// Simple implementation - in production use proper cron parser
	// This is a placeholder for the actual cron matching logic
	return false
}

func (s *Scheduler) addToDelayedQueue(task DelayedTask) {
	ctx := context.Background()

	score := float64(task.DelayUntil.Unix())
	s.redis.ZAdd(ctx, "delayed_queue", &redis.Z{
		Score:  score,
		Member: task.TaskID,
	})

	logger.WithFields(map[string]interface{}{
		"task_id":     task.TaskID,
		"delay_until": task.DelayUntil,
		"attempt":     task.Attempt,
	}).Info("Task added to delayed queue")
}

func (s *Scheduler) processDelayedTasks() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkDelayedTasks()
		}
	}
}

func (s *Scheduler) checkDelayedTasks() {
	ctx := context.Background()
	now := time.Now().Unix()

	// Get tasks ready to run
	tasks, err := s.redis.ZRangeByScore(ctx, "delayed_queue", &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", now),
	}).Result()

	if err != nil {
		logger.LogWithEvent(context.Background(), logger.LogEvent{
			Level:   "error",
			Message: "Failed to get delayed tasks",
			Error:   err.Error(),
		})
		return
	}

	logger.LogWithEvent(context.Background(), logger.LogEvent{
		Level:   "info",
		Message: "Processing delayed tasks",
		Fields:  map[string]interface{}{"count": len(tasks)},
	})

	for _, taskID := range tasks {
		// Remove from queue
		if err := s.redis.ZRem(ctx, "delayed_queue", taskID).Err(); err != nil {
			logger.LogWithEvent(context.Background(), logger.LogEvent{
				Level:   "error",
				Message: "Failed to remove task from queue",
				TaskID:  taskID,
				Error:   err.Error(),
			})
			continue
		}

		// Get task details
		var task models.Task
		if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
			logger.LogWithEvent(context.Background(), logger.LogEvent{
				Level:   "warn",
				Message: "Task not found in database",
				TaskID:  taskID,
				Error:   err.Error(),
			})
			continue
		}

		// Check if still should run
		if !task.IsActive {
			logger.LogWithEvent(context.Background(), logger.LogEvent{
				Level:    "info",
				Message:  "Skipping inactive delayed task",
				TaskID:   task.ID,
				TaskName: task.Name,
			})
			continue
		}

		// Try to execute
		canRun, delay, reason := s.checkResourceAvailability(&task)
		if canRun {
			jobID, err := s.executeTask(&task)
			if err != nil {
				logger.LogWithEvent(context.Background(), logger.LogEvent{
					Level:    "error",
					Message:  "Failed to execute delayed task",
					TaskID:   task.ID,
					TaskName: task.Name,
					Error:    err.Error(),
				})

				// Re-queue with backoff
				delayedTask := DelayedTask{
					TaskID:     task.ID,
					DelayUntil: time.Now().Add(30 * time.Second),
					Attempt:    1,
				}
				s.addToDelayedQueue(delayedTask)
			} else {
				logger.LogWithEvent(context.Background(), logger.LogEvent{
					Level:    "info",
					Message:  "Delayed task executed successfully",
					TaskID:   task.ID,
					TaskName: task.Name,
					JobID:    jobID,
				})
			}
		} else {
			logger.LogWithEvent(context.Background(), logger.LogEvent{
				Level:    "warn",
				Message:  "Delayed task delayed again",
				TaskID:   task.ID,
				TaskName: task.Name,
				Fields: map[string]interface{}{
					"reason": reason,
					"delay":  delay.String(),
				},
			})

			// Re-queue with increased delay
			delayedTask := DelayedTask{
				TaskID:     task.ID,
				DelayUntil: time.Now().Add(delay),
				Attempt:    2,
			}
			s.addToDelayedQueue(delayedTask)
		}
	}
}

func (s *Scheduler) scheduleTasks() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkAndScheduleTasks()
		}
	}
}

func (s *Scheduler) checkAndScheduleTasks() {
	var tasks []models.Task
	if err := s.db.Where("is_active = ?", true).Find(&tasks).Error; err != nil {
		logger.LogWithEvent(context.Background(), logger.LogEvent{
			Level:   "error",
			Message: "Failed to get active tasks",
			Error:   err.Error(),
			Fields:  map[string]interface{}{"error": err},
		})
		return
	}

	now := time.Now()

	for _, task := range tasks {
		if s.shouldRunNow(now, task.Schedule) {
			logger.LogWithEvent(context.Background(), logger.LogEvent{
				Level:    "info",
				Message:  "Task scheduled for execution",
				TaskID:   task.ID,
				TaskName: task.Name,
				Fields: map[string]interface{}{
					"schedule":    task.Schedule,
					"is_critical": task.IsCritical,
				},
			})

			// Async trigger task
			go func(t models.Task) {
				canRun, delay, reason := s.checkResourceAvailability(&t)
				if canRun || t.IsCritical {
					jobID, err := s.executeTask(&t)
					if err != nil {
						logger.LogWithEvent(context.Background(), logger.LogEvent{
							Level:    "error",
							Message:  "Failed to execute task",
							TaskID:   t.ID,
							TaskName: t.Name,
							Error:    err.Error(),
							Fields: map[string]interface{}{
								"job_id": jobID,
							},
						})
					} else {
						logger.LogWithEvent(context.Background(), logger.LogEvent{
							Level:    "info",
							Message:  "Task executed successfully",
							TaskID:   t.ID,
							TaskName: t.Name,
							JobID:    jobID,
							ExitCode: 0,
							Fields: map[string]interface{}{
								"critical": t.IsCritical,
							},
						})
					}
				} else {
					logger.LogWithEvent(context.Background(), logger.LogEvent{
						Level:    "warn",
						Message:  "Task delayed due to resource constraints",
						TaskID:   t.ID,
						TaskName: t.Name,
						Fields: map[string]interface{}{
							"delay":   delay.String(),
							"reason":  reason,
							"attempt": 1,
						},
					})

					delayedTask := DelayedTask{
						TaskID:     t.ID,
						DelayUntil: time.Now().Add(delay),
						Attempt:    1,
					}
					s.addToDelayedQueue(delayedTask)
				}
			}(task)
		}
	}
}

func (s *Scheduler) Start(port string) error {
	s.server = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go s.scheduleTasks()
	go s.processDelayedTasks()

	logger.WithField("port", port).Info("Scheduler starting")
	return s.server.ListenAndServe()
}

func (s *Scheduler) Stop(ctx context.Context) error {
	logger.Info("Shutting down Scheduler")
	close(s.stopCh)
	return s.server.Shutdown(ctx)
}

func initConfig() {
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "cron_user")
	viper.SetDefault("DB_PASSWORD", "cron_pass")
	viper.SetDefault("DB_NAME", "adaptive_cron")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("SERVER_PORT", "8082")
	viper.SetDefault("CHECK_INTERVAL", "10s")
	viper.SetDefault("METRIC_AGENT_URL", "http://localhost:8083")
	viper.SetDefault("WORKER_URL", "http://localhost:8084")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "text")

	viper.AutomaticEnv()
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
		viper.GetString("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT")),
		Password: "",
		DB:       0,
	})

	return client
}

func main() {
	initConfig()
	logger.Init(viper.GetString("LOG_LEVEL"), viper.GetString("LOG_FORMAT"))

	db, err := initDB()
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to connect to database")
	}

	redisClient := initRedis()
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.WithField("error", err).Fatal("Failed to connect to Redis")
	}

	checkInterval, err := time.ParseDuration(viper.GetString("CHECK_INTERVAL"))
	if err != nil {
		checkInterval = 10 * time.Second
	}

	scheduler := NewScheduler(
		db,
		redisClient,
		viper.GetString("METRIC_AGENT_URL"),
		viper.GetString("WORKER_URL"),
		checkInterval,
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := scheduler.Start(viper.GetString("SERVER_PORT")); err != nil && err != http.ErrServerClosed {
			logger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	<-sigChan

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := scheduler.Stop(shutdownCtx); err != nil {
		logger.WithField("error", err).Error("Failed to shutdown gracefully")
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
