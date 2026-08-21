package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Profile   string                 `mapstructure:"profile"`
	Database  DatabaseConfig         `mapstructure:"database"`
	Queue     QueueConfig            `mapstructure:"queue"`
	Logging   LoggingConfig          `mapstructure:"logging"`
	Server    ServerConfig           `mapstructure:"server"`
	Metrics   MetricsConfig          `mapstructure:"metrics"`
	Worker    WorkerConfig           `mapstructure:"worker"`
	Scheduler SchedulerConfig        `mapstructure:"scheduler"`
	Features  map[string]bool        `mapstructure:"features"`
	Custom    map[string]interface{} `mapstructure:"custom"`
}

type DatabaseConfig struct {
	Type        string            `mapstructure:"type"`
	DSN         string            `mapstructure:"dsn"`
	Host        string            `mapstructure:"host"`
	Port        int               `mapstructure:"port"`
	User        string            `mapstructure:"user"`
	Password    string            `mapstructure:"password"`
	DBName      string            `mapstructure:"dbname"`
	Replication ReplicationConfig `mapstructure:"replication"`
}

type QueueConfig struct {
	Type       string   `mapstructure:"type"`
	MaxSize    int      `mapstructure:"max_size"`
	Host       string   `mapstructure:"host"`
	Port       int      `mapstructure:"port"`
	DB         int      `mapstructure:"db"`
	Servers    []string `mapstructure:"servers"`
	Persistent bool     `mapstructure:"persistent"`
}

type LoggingConfig struct {
	Type          string         `mapstructure:"type"`
	Level         string         `mapstructure:"level"`
	Format        string         `mapstructure:"format"`
	File          string         `mapstructure:"file"`
	Rotation      RotationConfig `mapstructure:"rotation"`
	Host          string         `mapstructure:"host"`
	Port          int            `mapstructure:"port"`
	Database      string         `mapstructure:"database"`
	Table         string         `mapstructure:"table"`
	BatchSize     int            `mapstructure:"batch_size"`
	FlushInterval string         `mapstructure:"flush_interval"`
}

type ServerConfig struct {
	HTTPPort int `mapstructure:"http_port"`
	GRPCPort int `mapstructure:"grpc_port"`
	NATSPort int `mapstructure:"nats_port"`
}

type MetricsConfig struct {
	Type string `mapstructure:"type"`
	Port int    `mapstructure:"port"`
}

type WorkerConfig struct {
	MaxConcurrent int `mapstructure:"max_concurrent"`
	QueueSize     int `mapstructure:"queue_size"`
}

type SchedulerConfig struct {
	Type          string           `mapstructure:"type"`
	CheckInterval string           `mapstructure:"check_interval"`
	Leadership    LeadershipConfig `mapstructure:"leadership"`
}

type ReplicationConfig struct {
	Enabled    bool     `mapstructure:"enabled"`
	SlaveHosts []string `mapstructure:"slave_hosts"`
}

type RotationConfig struct {
	MaxSize    int `mapstructure:"max_size"`
	MaxAge     int `mapstructure:"max_age"`
	MaxBackups int `mapstructure:"max_backups"`
}

type LeadershipConfig struct {
	Type      string   `mapstructure:"type"`
	Endpoints []string `mapstructure:"endpoints"`
}

// Загрузка конфигурации
func LoadConfig(profile string) (*Config, error) {
	v := viper.New()

	// Установка имени конфига
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/adaptive-cron")

	// Переменные окружения
	v.SetEnvPrefix("ADAPTIVE_CRON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Загрузка базового конфига
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Если указан профиль, загружаем его
	if profile != "" {
		profileConfig := v.Sub(profile)
		if profileConfig != nil {
			v.MergeConfigMap(profileConfig.AllSettings())
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Создание DSN для разных БД
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "sqlite":
		return c.DSN
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.User, c.Password, c.DBName)
	default:
		return ""
	}
}
