package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"adaptive-cron/internal/common/models"
	"adaptive-cron/pkg/logger"
)

type TaskManager struct {
	db     *gorm.DB
	router *mux.Router
	server *http.Server
}

func NewTaskManager(db *gorm.DB) *TaskManager {
	tm := &TaskManager{
		db:     db,
		router: mux.NewRouter(),
	}
	tm.setupRoutes()
	return tm
}

func (tm *TaskManager) setupRoutes() {
	api := tm.router.PathPrefix("/api/v1/tasks").Subrouter()
	api.HandleFunc("", tm.createTask).Methods("POST")
	api.HandleFunc("", tm.listTasks).Methods("GET")
	api.HandleFunc("/{id}", tm.getTask).Methods("GET")
	api.HandleFunc("/{id}", tm.updateTask).Methods("PUT")
	api.HandleFunc("/{id}", tm.deleteTask).Methods("DELETE")
	api.HandleFunc("/{id}/toggle", tm.toggleTask).Methods("POST")

	tm.router.HandleFunc("/health", tm.healthCheck).Methods("GET")
}

func (tm *TaskManager) createTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task := &models.Task{
		Name:           req.Name,
		Command:        req.Command,
		Schedule:       req.Schedule,
		IsCritical:     req.IsCritical,
		MaxLoadScore:   req.MaxLoadScore,
		MaxCPUPercent:  req.MaxCPUPercent,
		MaxMemoryMB:    req.MaxMemoryMB,
		TimeoutSeconds: req.TimeoutSeconds,
		IsActive:       true,
	}

	if task.MaxLoadScore == 0 {
		task.MaxLoadScore = 0.7
	}
	if task.MaxCPUPercent == 0 {
		task.MaxCPUPercent = 80
	}
	if task.MaxMemoryMB == 0 {
		task.MaxMemoryMB = 1024
	}
	if task.TimeoutSeconds == 0 {
		task.TimeoutSeconds = 300
	}

	if err := task.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := tm.db.Create(task).Error; err != nil {
		logger.WithField("error", err).Error("Failed to create task")
		writeError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (tm *TaskManager) listTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []models.Task
	query := tm.db.Model(&models.Task{})

	if r.URL.Query().Get("active") == "true" {
		query = query.Where("is_active = ?", true)
	}
	if r.URL.Query().Get("critical") == "true" {
		query = query.Where("is_critical = ?", true)
	}

	if err := query.Find(&tasks).Error; err != nil {
		logger.WithField("error", err).Error("Failed to list tasks")
		writeError(w, http.StatusInternalServerError, "Failed to list tasks")
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (tm *TaskManager) getTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var task models.Task
	if err := tm.db.First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (tm *TaskManager) updateTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req models.UpdateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var task models.Task
	if err := tm.db.First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Command != nil {
		updates["command"] = *req.Command
	}
	if req.Schedule != nil {
		updates["schedule"] = *req.Schedule
	}
	if req.IsCritical != nil {
		updates["is_critical"] = *req.IsCritical
	}
	if req.MaxLoadScore != nil {
		updates["max_load_score"] = *req.MaxLoadScore
	}
	if req.MaxCPUPercent != nil {
		updates["max_cpu_percent"] = *req.MaxCPUPercent
	}
	if req.MaxMemoryMB != nil {
		updates["max_memory_mb"] = *req.MaxMemoryMB
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := tm.db.Model(&task).Updates(updates).Error; err != nil {
		logger.WithField("error", err).Error("Failed to update task")
		writeError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	tm.db.First(&task, "id = ?", id)
	writeJSON(w, http.StatusOK, task)
}

func (tm *TaskManager) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	result := tm.db.Delete(&models.Task{}, "id = ?", id)
	if result.Error != nil {
		logger.WithField("error", result.Error).Error("Failed to delete task")
		writeError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	if result.RowsAffected == 0 {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

func (tm *TaskManager) toggleTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var task models.Task
	if err := tm.db.First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}

	task.IsActive = !task.IsActive
	if err := tm.db.Save(&task).Error; err != nil {
		logger.WithField("error", err).Error("Failed to toggle task")
		writeError(w, http.StatusInternalServerError, "Failed to toggle task")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (tm *TaskManager) healthCheck(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := tm.db.DB()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Database connection error")
		return
	}

	if err := sqlDB.Ping(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "Database ping failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (tm *TaskManager) Start(port string) error {
	tm.server = &http.Server{
		Addr:         ":" + port,
		Handler:      tm.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.WithField("port", port).Info("Task Manager starting")
	return tm.server.ListenAndServe()
}

func (tm *TaskManager) Stop(ctx context.Context) error {
	logger.Info("Shutting down Task Manager")
	return tm.server.Shutdown(ctx)
}

func initConfig() {
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "cron_user")
	viper.SetDefault("DB_PASSWORD", "cron_pass")
	viper.SetDefault("DB_NAME", "adaptive_cron")
	viper.SetDefault("SERVER_PORT", "8081")
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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

func main() {
	initConfig()
	logger.Init(viper.GetString("LOG_LEVEL"), viper.GetString("LOG_FORMAT"))

	db, err := initDB()
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to connect to database")
	}

	// Auto migrate
	if err := db.AutoMigrate(&models.Task{}, &models.JobLog{}); err != nil {
		logger.WithField("error", err).Fatal("Failed to migrate database")
	}

	taskManager := NewTaskManager(db)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := taskManager.Start(viper.GetString("SERVER_PORT")); err != nil && err != http.ErrServerClosed {
			logger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := taskManager.Stop(shutdownCtx); err != nil {
		logger.WithField("error", err).Error("Failed to shutdown gracefully")
	}
}

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
