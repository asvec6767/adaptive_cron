package rest

import (
	"adaptive-cron/internal/apigateway/usecase"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	gatewayUseCase *usecase.GatewayUseCase
}

func NewHandler(gatewayUseCase *usecase.GatewayUseCase) *Handler {
	return &Handler{
		gatewayUseCase: gatewayUseCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Health
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/v1/gateway/health", h.GatewayHealth)

	// Service discovery
	mux.HandleFunc("GET /api/v1/gateway/services", h.ListServices)
	mux.HandleFunc("GET /api/v1/gateway/routes", h.ListRoutes)

	// Proxy all other requests
	mux.HandleFunc("/", h.Proxy)
}

func (h *Handler) Proxy(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Skip gateway internal routes
	if strings.HasPrefix(path, "/api/v1/gateway") {
		http.NotFound(w, r)
		return
	}

	// Find service for this path
	svc, err := h.gatewayUseCase.GetServiceByPath(path)
	if err != nil {
		writeError(w, http.StatusNotFound, "service not found for path: "+path)
		return
	}

	// Health check before proxying
	if err := h.gatewayUseCase.HealthCheck(r.Context(), svc.Name); err != nil {
		writeError(w, http.StatusServiceUnavailable, svc.Name+" is unavailable")
		return
	}

	// Get proxy
	proxy, err := h.gatewayUseCase.GetProxy(svc.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]string{"status": "healthy", "service": "api-gateway"}
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) GatewayHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := h.gatewayUseCase.ListServices()
	health := map[string]string{
		"gateway": "healthy",
	}

	allHealthy := true
	for _, svc := range services {
		err := h.gatewayUseCase.HealthCheck(r.Context(), svc.Name)
		if err != nil {
			health[svc.Name] = "unhealthy"
			allHealthy = false
		} else {
			health[svc.Name] = "healthy"
		}
	}

	if !allHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	writeJSON(w, http.StatusOK, health)
}

func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := h.gatewayUseCase.ListServices()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

func (h *Handler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	routes := h.gatewayUseCase.GetRoutes()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"routes": routes,
		"count":  len(routes),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
