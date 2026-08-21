package config

import (
	"adaptive-cron/internal/features"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Features    []string               `yaml:"features"`
	Config      map[string]interface{} `yaml:"config"`
}

type ProfileManager struct {
	profiles map[string]*Profile
	current  string
}

func NewProfileManager() *ProfileManager {
	return &ProfileManager{
		profiles: make(map[string]*Profile),
	}
}

// Загрузка профиля из файла
func (pm *ProfileManager) LoadProfile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return err
	}

	pm.profiles[profile.Name] = &profile
	return nil
}

// Загрузка встроенных профилей
func (pm *ProfileManager) LoadBuiltinProfiles() {
	// Profile 1: Lite
	pm.profiles["lite"] = &Profile{
		Name:        "lite",
		Description: "Lightweight version for small websites and individual users",
		Features: []string{
			"db.sqlite",
			"queue.memory",
			"log.file",
			"transport.rest",
			"metrics.inmemory",
			"monitoring.health",
		},
		Config: map[string]interface{}{
			"database": map[string]interface{}{
				"type": "sqlite",
				"path": "./data/adaptive-cron.db",
			},
			"queue": map[string]interface{}{
				"type":     "memory",
				"max_size": 100,
			},
			"logging": map[string]interface{}{
				"type":  "file",
				"file":  "./logs/adaptive-cron.log",
				"level": "info",
			},
			"server": map[string]interface{}{
				"http_port": 8080,
			},
		},
	}

	// Profile 2: Standard
	pm.profiles["standard"] = &Profile{
		Name:        "standard",
		Description: "Standard version with full logging and monitoring",
		Features: []string{
			"db.postgres",
			"queue.redis",
			"log.file",
			"log.json",
			"log.structured",
			"transport.rest",
			"transport.grpc",
			"metrics.prometheus",
			"worker.pool",
			"task.history",
			"monitoring.health",
			"performance.caching",
			"monitoring.profiling",
		},
		Config: map[string]interface{}{
			"database": map[string]interface{}{
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"user":     "cron_user",
				"password": "cron_pass",
				"dbname":   "adaptive_cron",
			},
			"queue": map[string]interface{}{
				"type": "redis",
				"host": "localhost",
				"port": 6379,
				"db":   0,
			},
			"logging": map[string]interface{}{
				"type":   "file",
				"format": "json",
				"file":   "./logs/adaptive-cron.log",
				"level":  "info",
				"rotation": map[string]interface{}{
					"max_size":    100,
					"max_age":     30,
					"max_backups": 5,
				},
			},
			"server": map[string]interface{}{
				"http_port": 8080,
				"grpc_port": 9090,
			},
			"metrics": map[string]interface{}{
				"port": 9091,
			},
			"worker": map[string]interface{}{
				"max_concurrent": 10,
			},
		},
	}

	// Profile 3: Enterprise
	pm.profiles["enterprise"] = &Profile{
		Name:        "enterprise",
		Description: "Enterprise version with distributed architecture",
		Features: []string{
			"db.postgres",
			"db.replication",
			"queue.nats",
			"queue.persistent",
			"log.clickhouse",
			"log.structured",
			"transport.grpc",
			"transport.nats",
			"metrics.prometheus",
			"service.discovery",
			"scheduler.distributed",
			"worker.pool",
			"task.dependencies",
			"task.history",
			"monitoring.tracing",
			"monitoring.health",
			"performance.batching",
			"performance.caching",
			"performance.compression",
			"monitoring.profiling",
		},
		Config: map[string]interface{}{
			"database": map[string]interface{}{
				"type":     "postgres",
				"host":     "postgres-master",
				"port":     5432,
				"user":     "cron_user",
				"password": "cron_pass",
				"dbname":   "adaptive_cron",
				"replication": map[string]interface{}{
					"enabled":     true,
					"slave_hosts": []string{"postgres-slave1", "postgres-slave2"},
				},
			},
			"queue": map[string]interface{}{
				"type":       "nats",
				"servers":    []string{"nats://nats1:4222", "nats://nats2:4222"},
				"cluster_id": "adaptive-cron",
				"persistent": true,
			},
			"logging": map[string]interface{}{
				"type":           "clickhouse",
				"host":           "clickhouse",
				"port":           9000,
				"database":       "adaptive_cron",
				"table":          "logs",
				"batch_size":     100,
				"flush_interval": "5s",
			},
			"server": map[string]interface{}{
				"http_port": 8080,
				"grpc_port": 9090,
				"nats_port": 4222,
			},
			"metrics": map[string]interface{}{
				"port": 9091,
			},
			"worker": map[string]interface{}{
				"max_concurrent": 50,
				"queue_size":     1000,
			},
			"scheduler": map[string]interface{}{
				"type":           "distributed",
				"check_interval": "10s",
				"leadership": map[string]interface{}{
					"type":      "etcd",
					"endpoints": []string{"etcd1:2379", "etcd2:2379"},
				},
			},
		},
	}

	// Profile 4: Cloud
	pm.profiles["cloud"] = &Profile{
		Name:        "cloud",
		Description: "Cloud-native version with full observability",
		Features: []string{
			"db.postgres",
			"db.replication",
			"queue.nats",
			"queue.persistent",
			"log.clickhouse",
			"log.structured",
			"log.elk",
			"transport.grpc",
			"transport.nats",
			"metrics.prometheus",
			"service.discovery",
			"scheduler.distributed",
			"worker.pool",
			"task.dependencies",
			"task.history",
			"monitoring.tracing",
			"monitoring.health",
			"performance.batching",
			"performance.caching",
			"performance.compression",
			"monitoring.profiling",
		},
		Config: map[string]interface{}{
			// Cloud configuration
		},
	}
}

// Применить профиль
func (pm *ProfileManager) ApplyProfile(name string) (*Profile, error) {
	profile, ok := pm.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %s not found", name)
	}

	pm.current = name

	// Включить feature flags
	for _, feature := range profile.Features {
		features.Enable(feature)
	}

	return profile, nil
}
