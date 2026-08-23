package usecase

import (
	"adaptive-cron/internal/apigateway/domain"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type GatewayUseCase struct {
	mu         sync.RWMutex
	services   map[string]*domain.ServiceConfig
	proxies    map[string]*httputil.ReverseProxy
	httpClient *http.Client
	routes     map[string]string // path -> service name
}

func NewGatewayUseCase() *GatewayUseCase {
	return &GatewayUseCase{
		services: make(map[string]*domain.ServiceConfig),
		proxies:  make(map[string]*httputil.ReverseProxy),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		routes: make(map[string]string),
	}
}

func (g *GatewayUseCase) RegisterService(config *domain.ServiceConfig) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	targetURL, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("invalid service URL: %w", err)
	}

	g.services[config.Name] = config
	g.proxies[config.Name] = httputil.NewSingleHostReverseProxy(targetURL)

	// Register routes based on API path
	if config.APIPath != "" {
		g.routes[config.APIPath] = config.Name
	}

	return nil
}

func (g *GatewayUseCase) GetProxy(serviceName string) (*httputil.ReverseProxy, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	proxy, exists := g.proxies[serviceName]
	if !exists {
		return nil, domain.ErrServiceNotFound
	}

	return proxy, nil
}

func (g *GatewayUseCase) GetServiceByName(name string) (*domain.ServiceConfig, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	svc, exists := g.services[name]
	if !exists {
		return nil, domain.ErrServiceNotFound
	}

	return svc, nil
}

func (g *GatewayUseCase) GetServiceByPath(path string) (*domain.ServiceConfig, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for prefix, name := range g.routes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			svc, exists := g.services[name]
			if !exists {
				return nil, domain.ErrServiceNotFound
			}
			return svc, nil
		}
	}

	return nil, domain.ErrRouteNotFound
}

func (g *GatewayUseCase) HealthCheck(ctx context.Context, serviceName string) error {
	svc, err := g.GetServiceByName(serviceName)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", svc.URL+svc.HealthPath, nil)
	if err != nil {
		return err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.ErrServiceUnavailable
	}

	return nil
}

func (g *GatewayUseCase) ListServices() []*domain.ServiceConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()

	services := make([]*domain.ServiceConfig, 0, len(g.services))
	for _, svc := range g.services {
		services = append(services, svc)
	}
	return services
}

func (g *GatewayUseCase) GetRoutes() map[string]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.routes
}
