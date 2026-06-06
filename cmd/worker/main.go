package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "sync"
    "syscall"
    "time"
    
    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/spf13/viper"
    
    "adaptive-cron/internal/common/models"
    "adaptive-cron/pkg/logger"
)

type Job struct {
    ID        string
    TaskID    string
    Command   string
    Timeout   time.Duration
    StartTime time.Time
    EndTime   time.Time
    ExitCode  int
    Output    string
    Error     string
    Done      chan bool
}

type Worker struct {
    mu            sync.RWMutex
    jobs          map[string]*Job
    maxConcurrent int
    semaphore     chan struct{}
    router        *mux.Router
    server        *http.Server
}

func NewWorker(maxConcurrent int) *Worker {
    w := &Worker{
        jobs:          make(map[string]*Job),
        maxConcurrent: maxConcurrent,
        semaphore:     make(chan struct{}, maxConcurrent),
        router:        mux.NewRouter(),
    }
    w.setupRoutes()
    return w
}

func (w *Worker) setupRoutes() {
    api := w.router.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/execute", w.executeJob).Methods("POST")
    api.HandleFunc("/status/{jobId}", w.getJobStatus).Methods("GET")
    api.HandleFunc("/jobs", w.listJobs).Methods("GET")
    
    w.router.HandleFunc("/health", w.healthCheck).Methods("GET")
}

type ExecuteRequest struct {
    TaskID  string `json:"task_id"`
    Command string `json:"command"`
    Timeout int    `json:"timeout"`
}

type ExecuteResponse struct {
    JobID string `json:"job_id"`
}

func (w *Worker) executeJob(rw http.ResponseWriter, r *http.Request) {
    var req ExecuteRequest
    if err := decodeJSON(r, &req); err != nil {
        writeError(rw, http.StatusBadRequest, err.Error())
        return
    }
    
    if req.Command == "" {
        writeError(rw, http.StatusBadRequest, "command is required")
        return
    }
    
    jobID := uuid.New().String()
    timeout := time.Duration(req.Timeout) * time.Second
    if timeout == 0 {
        timeout = 300 * time.Second
    }
    
    job := &Job{
        ID:      jobID,
        TaskID:  req.TaskID,
        Command: req.Command,
        Timeout: timeout,
        Done:    make(chan bool),
    }
    
    w.mu.Lock()
    w.jobs[jobID] = job
    w.mu.Unlock()
    
    go w.runJob(job)
    
    writeJSON(rw, http.StatusAccepted, ExecuteResponse{JobID: jobID})
}

func (w *Worker) runJob(job *Job) {
    // Acquire semaphore
    select {
    case w.semaphore <- struct{}{}:
        defer func() { <-w.semaphore }()
    case <-time.After(30 * time.Second):
        job.Error = "Failed to acquire execution slot"
        job.Done <- true
        return
    }
    
    job.StartTime = time.Now()
    logger.WithFields(map[string]interface{}{
        "job_id":  job.ID,
        "command": job.Command,
    }).Info("Starting job execution")
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
    defer cancel()
    
    // Execute command
    cmd := exec.CommandContext(ctx, "sh", "-c", job.Command)
    output, err := cmd.CombinedOutput()
    
    job.EndTime = time.Now()
    job.Output = string(output)
    
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            job.Error = "Execution timeout"
            job.ExitCode = -1
        } else if exitErr, ok := err.(*exec.ExitError); ok {
            job.ExitCode = exitErr.ExitCode()
            job.Error = err.Error()
        } else {
            job.ExitCode = -1
            job.Error = err.Error()
        }
        logger.WithFields(map[string]interface{}{
            "job_id":    job.ID,
            "exit_code": job.ExitCode,
            "error":     job.Error,
        }).Error("Job execution failed")
    } else {
        job.ExitCode = 0
        logger.WithFields(map[string]interface{}{
            "job_id":    job.ID,
            "exit_code": 0,
            "duration":  job.EndTime.Sub(job.StartTime),
        }).Info("Job execution completed successfully")
    }
    
    job.Done <- true
}

func (w *Worker) getJobStatus(rw http.ResponseWriter, r *http.Request) {
    jobID := mux.Vars(r)["jobId"]
    
    w.mu.RLock()
    job, exists := w.jobs[jobID]
    w.mu.RUnlock()
    
    if !exists {
        writeError(rw, http.StatusNotFound, "Job not found")
        return
    }
    
    status := map[string]interface{}{
        "job_id":     job.ID,
        "task_id":    job.TaskID,
        "command":    job.Command,
        "start_time": job.StartTime,
        "end_time":   job.EndTime,
        "exit_code":  job.ExitCode,
        "output":     job.Output,
        "error":      job.Error,
    }
    
    writeJSON(rw, http.StatusOK, status)
}

func (w *Worker) listJobs(rw http.ResponseWriter, r *http.Request) {
    w.mu.RLock()
    defer w.mu.RUnlock()
    
    jobs := make([]map[string]interface{}, 0, len(w.jobs))
    for _, job := range w.jobs {
        jobs = append(jobs, map[string]interface{}{
            "job_id":     job.ID,
            "task_id":    job.TaskID,
            "start_time": job.StartTime,
            "end_time":   job.EndTime,
            "exit_code":  job.ExitCode,
        })
    }
    
    writeJSON(rw, http.StatusOK, map[string]interface{}{
        "jobs": jobs,
        "total": len(jobs),
    })
}

func (w *Worker) healthCheck(rw http.ResponseWriter, r *http.Request) {
    w.mu.RLock()
    activeJobs := len(w.semaphore)
    w.mu.RUnlock()
    
    writeJSON(rw, http.StatusOK, map[string]interface{}{
        "status":         "healthy",
        "active_jobs":    activeJobs,
        "max_concurrent": w.maxConcurrent,
    })
}

func (w *Worker) Start(port string) error {
    w.server = &http.Server{
        Addr:         ":" + port,
        Handler:      w.router,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    }
    
    logger.WithField("port", port).Info("Worker starting")
    return w.server.ListenAndServe()
}

func (w *Worker) Stop(ctx context.Context) error {
    logger.Info("Shutting down Worker")
    
    // Wait for active jobs to complete
    w.mu.RLock()
    activeCount := len(w.semaphore)
    w.mu.RUnlock()
    
    if activeCount > 0 {
        logger.WithField("active_jobs", activeCount).Info("Waiting for active jobs to complete")
        time.Sleep(5 * time.Second)
    }
    
    return w.server.Shutdown(ctx)
}

func initConfig() {
    viper.SetDefault("SERVER_PORT", "8084")
    viper.SetDefault("MAX_CONCURRENT", "5")
    viper.SetDefault("LOG_LEVEL", "info")
    viper.SetDefault("LOG_FORMAT", "text")
    
    viper.AutomaticEnv()
}

func main() {
    initConfig()
    logger.Init(viper.GetString("LOG_LEVEL"), viper.GetString("LOG_FORMAT"))
    
    maxConcurrent := viper.GetInt("MAX_CONCURRENT")
    if maxConcurrent <= 0 {
        maxConcurrent = 5
    }
    
    worker := NewWorker(maxConcurrent)
    
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
    
    go func() {
        if err := worker.Start(viper.GetString("SERVER_PORT")); err != nil && err != http.ErrServerClosed {
            logger.WithField("error", err).Fatal("Failed to start server")
        }
    }()
    
    <-ctx.Done()
    
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := worker.Stop(shutdownCtx); err != nil {
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