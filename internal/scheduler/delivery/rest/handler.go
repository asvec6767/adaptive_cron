package rest

import (
	"adaptive-cron/internal/scheduler/usecase"
	"encoding/json"
	"net/http"
)

type Handler struct {
	schedulerUseCase *usecase.SchedulerUseCase
}

func NewHandler(schedulerUseCase *usecase.SchedulerUseCase) *Handler {
	return &Handler{
		schedulerUseCase: schedulerUseCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/scheduler/next", h.GetNextTasks)
	mux.HandleFunc("POST /api/v1/scheduler/trigger/{id}", h.TriggerTask)
	mux.HandleFunc("GET /api/v1/scheduler/queue", h.GetDelayedQueue)
	mux.HandleFunc("DELETE /api/v1/scheduler/queue/{id}", h.RemoveFromQueue)
	mux.HandleFunc("GET /health", h.Health)
}

func (h *Handler) GetNextTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Placeholder response
	response := map[string]interface{}{
		"tasks": []interface{}{},
		"count": 0,
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) TriggerTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.PathValue("id")
	force := r.URL.Query().Get("force") == "true"

	err := h.schedulerUseCase.TriggerTask(r.Context(), taskID, force)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "triggered"})
}

func (h *Handler) GetDelayedQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := h.schedulerUseCase.GetDelayedQueue(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func (h *Handler) RemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.PathValue("id")
	err := h.schedulerUseCase.RemoveFromQueue(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
