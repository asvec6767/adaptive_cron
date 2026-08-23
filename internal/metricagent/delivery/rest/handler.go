package rest

import (
	"adaptive-cron/internal/metricagent/usecase"
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase *usecase.MetricAgentUseCase
}

func NewHandler(useCase *usecase.MetricAgentUseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/metrics/current", h.GetCurrentMetrics)
	mux.HandleFunc("GET /api/v1/metrics/history", h.GetMetricsHistory)
	mux.HandleFunc("GET /api/v1/metrics/load", h.GetLoadScore)
	mux.HandleFunc("GET /api/v1/metrics/health", h.Health)
	mux.HandleFunc("GET /health", h.Health)
}

func (h *Handler) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := h.useCase.GetCurrentMetrics(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

func (h *Handler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.useCase.GetMetricsHistory(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": history,
		"total":   len(history),
	})
}

func (h *Handler) GetLoadScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	score, err := h.useCase.GetLoadScore(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]float64{
		"load_score": score,
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := h.useCase.GetCurrentMetrics(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "metrics not available")
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
