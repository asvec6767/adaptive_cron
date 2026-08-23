package domain

type ServiceConfig struct {
	Name       string
	URL        string
	Port       int
	HealthPath string
	APIPath    string
}

type GatewayConfig struct {
	Port          int
	Services      []ServiceConfig
	EnableGRPC    bool
	EnableMetrics bool
}
