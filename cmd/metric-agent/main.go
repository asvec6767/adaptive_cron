package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/viper"

	"adaptive-cron/internal/common/models"
	"adaptive-cron/pkg/logger"
)

type MetricAgent struct {
	mu              sync.RWMutex
	currentMetrics  *models.SystemMetrics
	history         []models.SystemMetrics
	maxHistorySize  int
	collectInterval time.Duration
	router          *mux.Router
	server          *http.Server
	stopCh          chan struct{}
}

func NewMetricAgent(collectInterval time.Duration, maxHistorySize int) *MetricAgent {
	ma := &MetricAgent{
		currentMetrics:  &models.SystemMetrics{},
		history:         make([]models.SystemMetrics, 0, maxHistorySize),
		maxHistorySize:  maxHistorySize,
		collectInterval: collectInterval,
		router:          mux.NewRouter(),
		stopCh:          make(chan struct{}),
	}
	ma.setupRoutes()
	return ma
}

func (ma *MetricAgent) setupRoutes() {
	ma.router.HandleFunc("/api/v1/metrics/current", ma.getCurrentMetrics).Methods("GET")
	ma.router.HandleFunc("/api/v1/metrics/history", ma.getMetricsHistory).Methods("GET")
	ma.router.HandleFunc("/api/v1/metrics/health", ma.healthCheck).Methods("GET")
	ma.router.HandleFunc("/api/v1/metrics/load", ma.getLoadScore).Methods("GET")
}

func (ma *MetricAgent) getCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	writeJSON(w, http.StatusOK, ma.currentMetrics)
}

func (ma *MetricAgent) getMetricsHistory(w http.ResponseWriter, r *http.Request) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	start := len(ma.history) - limit
	if start < 0 {
		start = 0
	}

	history := ma.history[start:]
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": history,
		"total":   len(history),
	})
}

func (ma *MetricAgent) getLoadScore(w http.ResponseWriter, r *http.Request) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	score := ma.currentMetrics.GetLoadScore()
	writeJSON(w, http.StatusOK, map[string]float64{
		"load_score": score,
	})
}

func (ma *MetricAgent) healthCheck(w http.ResponseWriter, r *http.Request) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	if ma.currentMetrics.Timestamp.IsZero() {
		writeError(w, http.StatusServiceUnavailable, "Metrics not yet collected")
		return
	}

	if time.Since(ma.currentMetrics.Timestamp) > ma.collectInterval*2 {
		writeError(w, http.StatusServiceUnavailable, "Metrics are stale")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (ma *MetricAgent) collectMetrics() {
	ticker := time.NewTicker(ma.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ma.stopCh:
			logger.Info("Stopping metrics collection")
			return
		case <-ticker.C:
			ma.collectOnce()
		}
	}
}

func (ma *MetricAgent) collectOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metrics := &models.SystemMetrics{
		Timestamp: time.Now(),
	}

	// CPU Percent
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	} else {
		logger.WithField("error", err).Warn("Failed to collect CPU metrics")
	}

	// Load Average
	loadAvg, err := load.AvgWithContext(ctx)
	if err == nil {
		metrics.LoadAvg1 = loadAvg.Load1
		metrics.LoadAvg5 = loadAvg.Load5
		metrics.LoadAvg15 = loadAvg.Load15
	} else {
		logger.WithField("error", err).Warn("Failed to collect load average")
	}

	// Memory
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		metrics.MemoryUsed = memInfo.Used
		metrics.MemoryTotal = memInfo.Total
		metrics.MemoryPercent = memInfo.UsedPercent
	} else {
		logger.WithField("error", err).Warn("Failed to collect memory metrics")
	}

	// Swap
	swapInfo, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		metrics.SwapUsed = swapInfo.Used
		metrics.SwapTotal = swapInfo.Total
		metrics.SwapPercent = swapInfo.UsedPercent
	} else {
		logger.WithField("error", err).Warn("Failed to collect swap metrics")
	}

	ma.mu.Lock()
	ma.currentMetrics = metrics

	ma.history = append(ma.history, *metrics)
	if len(ma.history) > ma.maxHistorySize {
		ma.history = ma.history[1:]
	}
	ma.mu.Unlock()

	logger.WithFields(logger.WithFields(map[string]interface{}{
		"cpu":    fmt.Sprintf("%.2f%%", metrics.CPUPercent),
		"memory": fmt.Sprintf("%.2f%%", metrics.MemoryPercent),
		"load":   fmt.Sprintf("%.2f", metrics.LoadAvg1),
		"score":  fmt.Sprintf("%.2f", metrics.GetLoadScore()),
	})).Debug("Metrics collected")
}

func (ma *MetricAgent) Start(port string) error {
	ma.server = &http.Server{
		Addr:         ":" + port,
		Handler:      ma.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go ma.collectMetrics()

	logger.WithField("port", port).Info("Metric Agent starting")
	return ma.server.ListenAndServe()
}

func (ma *MetricAgent) Stop(ctx context.Context) error {
	logger.Info("Shutting down Metric Agent")
	close(ma.stopCh)
	return ma.server.Shutdown(ctx)
}

func initConfig() {
	viper.SetDefault("SERVER_PORT", "8083")
	viper.SetDefault("COLLECT_INTERVAL", "5s")
	viper.SetDefault("MAX_HISTORY_SIZE", 1000)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "text")

	viper.AutomaticEnv()
}

func main() {
	initConfig()
	logger.Init(viper.GetString("LOG_LEVEL"), viper.GetString("LOG_FORMAT"))

	collectInterval, err := time.ParseDuration(viper.GetString("COLLECT_INTERVAL"))
	if err != nil {
		logger.WithField("error", err).Fatal("Invalid collect interval")
	}

	maxHistorySize := viper.GetInt("MAX_HISTORY_SIZE")

	metricAgent := NewMetricAgent(collectInterval, maxHistorySize)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := metricAgent.Start(viper.GetString("SERVER_PORT")); err != nil && err != http.ErrServerClosed {
			logger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := metricAgent.Stop(shutdownCtx); err != nil {
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
