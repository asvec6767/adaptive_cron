package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"adaptive-cron/internal/common/restclient"
	"adaptive-cron/pkg/logger"
)

type APIGateway struct {
	router         *mux.Router
	server         *http.Server
	taskManagerURL string
	schedulerURL   string
	metricAgentURL string
	workerURL      string

	taskManagerProxy *httputil.ReverseProxy
	schedulerProxy   *httputil.ReverseProxy
	metricAgentProxy *httputil.ReverseProxy
	workerProxy      *httputil.ReverseProxy
}

func NewAPIGateway(taskManagerURL, schedulerURL, metricAgentURL, workerURL string) *APIGateway {
	gw := &APIGateway{
		router:         mux.NewRouter(),
		taskManagerURL: taskManagerURL,
		schedulerURL:   schedulerURL,
		metricAgentURL: metricAgentURL,
		workerURL:      workerURL,
	}

	gw.setupProxies()
	gw.setupRoutes()
	return gw
}

func (gw *APIGateway) setupProxies() {
	// Task Manager proxy
	taskManagerTarget, _ := url.Parse(gw.taskManagerURL)
	gw.taskManagerProxy = httputil.NewSingleHostReverseProxy(taskManagerTarget)

	// Scheduler proxy
	schedulerTarget, _ := url.Parse(gw.schedulerURL)
	gw.schedulerProxy = httputil.NewSingleHostReverseProxy(schedulerTarget)

	// Metric Agent proxy
	metricAgentTarget, _ := url.Parse(gw.metricAgentURL)
	gw.metricAgentProxy = httputil.NewSingleHostReverseProxy(metricAgentTarget)

	// Worker proxy
	workerTarget, _ := url.Parse(gw.workerURL)
	gw.workerProxy = httputil.NewSingleHostReverseProxy(workerTarget)
}

func (gw *APIGateway) setupRoutes() {
	// Health check
	gw.router.HandleFunc("/health", gw.healthCheck).Methods("GET")
	gw.router.HandleFunc("/", gw.getInfo).Methods("GET")

	// Task routes
	gw.router.PathPrefix("/api/v1/tasks").Handler(gw.proxyHandler(gw.taskManagerProxy))

	// Scheduler routes
	gw.router.PathPrefix("/api/v1/scheduler").Handler(gw.proxyHandler(gw.schedulerProxy))

	// Metrics routes
	gw.router.PathPrefix("/api/v1/metrics").Handler(gw.proxyHandler(gw.metricAgentProxy))

	// Worker routes
	gw.router.PathPrefix("/api/v1/execute").Handler(gw.proxyHandler(gw.workerProxy))
	gw.router.PathPrefix("/api/v1/status").Handler(gw.proxyHandler(gw.workerProxy))
	gw.router.PathPrefix("/api/v1/jobs").Handler(gw.proxyHandler(gw.workerProxy))
}

func (gw *APIGateway) proxyHandler(proxy *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Debug("Proxying request")

		proxy.ServeHTTP(w, r)
	})
}

func (gw *APIGateway) healthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]string{
		"status":  "healthy",
		"service": "api-gateway",
	}

	// Check backend services
	services := map[string]string{
		"task_manager": gw.taskManagerURL + "/health",
		"scheduler":    gw.schedulerURL + "/health",
		"metric_agent": gw.metricAgentURL + "/health",
		"worker":       gw.workerURL + "/health",
	}

	client := restclient.NewClient("", 2*time.Second)
	allHealthy := true

	for name, url := range services {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var result map[string]string
		err := client.Get(ctx, url, &result)
		if err != nil || result["status"] != "healthy" {
			allHealthy = false
			health[name] = "unhealthy"
		} else {
			health[name] = "healthy"
		}
	}

	if !allHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	writeJSON(w, http.StatusOK, health)
}

func (gw *APIGateway) getInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"service": "Adaptive Cron API Gateway",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"tasks":     "/api/v1/tasks",
			"scheduler": "/api/v1/scheduler",
			"metrics":   "/api/v1/metrics",
			"execute":   "/api/v1/execute",
			"jobs":      "/api/v1/jobs",
		},
	})
}

func (gw *APIGateway) Start(port string) error {
	gw.server = &http.Server{
		Addr:         ":" + port,
		Handler:      gw.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.WithField("port", port).Info("API Gateway starting")
	return gw.server.ListenAndServe()
}

func (gw *APIGateway) Stop(ctx context.Context) error {
	logger.Info("Shutting down API Gateway")
	return gw.server.Shutdown(ctx)
}

func initConfig() {
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("TASK_MANAGER_URL", "http://localhost:8081")
	viper.SetDefault("SCHEDULER_URL", "http://localhost:8082")
	viper.SetDefault("METRIC_AGENT_URL", "http://localhost:8083")
	viper.SetDefault("WORKER_URL", "http://localhost:8084")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "text")

	viper.AutomaticEnv()
}

func main() {
	initConfig()
	logger.Init(viper.GetString("LOG_LEVEL"), viper.GetString("LOG_FORMAT"))

	gateway := NewAPIGateway(
		viper.GetString("TASK_MANAGER_URL"),
		viper.GetString("SCHEDULER_URL"),
		viper.GetString("METRIC_AGENT_URL"),
		viper.GetString("WORKER_URL"),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := gateway.Start(viper.GetString("SERVER_PORT")); err != nil && err != http.ErrServerClosed {
			logger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := gateway.Stop(shutdownCtx); err != nil {
		logger.WithField("error", err).Error("Failed to shutdown gracefully")
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
