package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig(profile string, configFile string) (*Config, error) {
	v := viper.New()

	// Set config name and type
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/adaptive-cron")
	}

	// Environment variables with prefix
	v.SetEnvPrefix("ADAPTIVE_CRON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read base config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Load profile defaults if profile specified
	if profile != "" {
		if err := loadProfileDefaults(v, profile); err != nil {
			return nil, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	config.Profile = profile
	return &config, nil
}

func loadProfileDefaults(v *viper.Viper, profile string) error {
	defaults := getProfileDefaults(profile)
	if defaults == nil {
		return fmt.Errorf("profile %s not found", profile)
	}

	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	return nil
}

func getProfileDefaults(profile string) map[string]interface{} {
	switch profile {
	case "lite":
		return map[string]interface{}{
			"database.type":            "sqlite",
			"database.dsn":             "./data/adaptive-cron.db",
			"queue.type":               "memory",
			"queue.max_size":           100,
			"logging.type":             "file",
			"logging.level":            "info",
			"logging.format":           "text",
			"logging.file":             "./logs/adaptive-cron.log",
			"server.http_port":         8080,
			"worker.max_concurrent":    3,
			"scheduler.check_interval": "10s",
			"scheduler.type":           "local",
			"metrics.type":             "basic",
			"features": []string{
				"db.sqlite",
				"queue.memory",
				"log.file",
				"transport.inmemory",
				"metrics.basic",
			},
		}
	case "standard":
		return map[string]interface{}{
			"database.type":                "postgres",
			"database.host":                "localhost",
			"database.port":                5432,
			"database.user":                "cron_user",
			"database.password":            "cron_pass",
			"database.dbname":              "adaptive_cron",
			"queue.type":                   "redis",
			"queue.host":                   "localhost",
			"queue.port":                   6379,
			"queue.db":                     0,
			"logging.type":                 "file",
			"logging.level":                "info",
			"logging.format":               "json",
			"logging.file":                 "./logs/adaptive-cron.log",
			"logging.rotation.max_size":    100,
			"logging.rotation.max_age":     30,
			"logging.rotation.max_backups": 5,
			"server.http_port":             8080,
			"server.grpc_port":             9090,
			"worker.max_concurrent":        10,
			"scheduler.check_interval":     "10s",
			"scheduler.type":               "local",
			"metrics.type":                 "prometheus",
			"metrics.port":                 9091,
			"features": []string{
				"db.postgres",
				"queue.redis",
				"log.file",
				"log.json",
				"transport.rest",
				"transport.grpc",
				"metrics.prometheus",
			},
		}
	case "enterprise":
		return map[string]interface{}{
			"database.type":            "postgres",
			"database.host":            "postgres-master",
			"database.port":            5432,
			"database.user":            "cron_user",
			"database.password":        "cron_pass",
			"database.dbname":          "adaptive_cron",
			"queue.type":               "redis",
			"queue.host":               "redis",
			"queue.port":               6379,
			"queue.db":                 0,
			"logging.type":             "clickhouse",
			"logging.level":            "info",
			"logging.format":           "json",
			"logging.host":             "clickhouse",
			"logging.port":             9000,
			"logging.database":         "adaptive_cron",
			"logging.table":            "logs",
			"logging.batch_size":       100,
			"logging.flush_interval":   "5s",
			"server.http_port":         8080,
			"server.grpc_port":         9090,
			"worker.max_concurrent":    50,
			"worker.queue_size":        1000,
			"scheduler.check_interval": "10s",
			"scheduler.type":           "distributed",
			"metrics.type":             "prometheus",
			"metrics.port":             9091,
			"features": []string{
				"db.postgres",
				"db.replication",
				"queue.redis",
				"queue.persistent",
				"log.clickhouse",
				"log.structured",
				"transport.grpc",
				"transport.nats",
				"metrics.prometheus",
				"service.discovery",
				"scheduler.distributed",
			},
		}
	default:
		return nil
	}
}
