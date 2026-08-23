package config

import (
	"fmt"
	"time"
)

type Config struct {
	Profile   string          `yaml:"profile" mapstructure:"profile"`
	Database  DatabaseConfig  `yaml:"database" mapstructure:"database"`
	Queue     QueueConfig     `yaml:"queue" mapstructure:"queue"`
	Logging   LoggingConfig   `yaml:"logging" mapstructure:"logging"`
	Server    ServerConfig    `yaml:"server" mapstructure:"server"`
	Worker    WorkerConfig    `yaml:"worker" mapstructure:"worker"`
	Scheduler SchedulerConfig `yaml:"scheduler" mapstructure:"scheduler"`
	Metrics   MetricsConfig   `yaml:"metrics" mapstructure:"metrics"`
	Features  []string        `yaml:"features" mapstructure:"features"`
}

type DatabaseConfig struct {
	Type     string `yaml:"type" mapstructure:"type"`
	DSN      string `yaml:"dsn" mapstructure:"dsn"`
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"dbname" mapstructure:"dbname"`
}

type QueueConfig struct {
	Type    string   `yaml:"type" mapstructure:"type"`
	MaxSize int      `yaml:"max_size" mapstructure:"max_size"`
	Host    string   `yaml:"host" mapstructure:"host"`
	Port    int      `yaml:"port" mapstructure:"port"`
	DB      int      `yaml:"db" mapstructure:"db"`
	Servers []string `yaml:"servers" mapstructure:"servers"`
}

type LoggingConfig struct {
	Type          string         `yaml:"type" mapstructure:"type"`
	Level         string         `yaml:"level" mapstructure:"level"`
	Format        string         `yaml:"format" mapstructure:"format"`
	File          string         `yaml:"file" mapstructure:"file"`
	Rotation      RotationConfig `yaml:"rotation" mapstructure:"rotation"`
	Host          string         `yaml:"host" mapstructure:"host"`
	Port          int            `yaml:"port" mapstructure:"port"`
	Database      string         `yaml:"database" mapstructure:"database"`
	Table         string         `yaml:"table" mapstructure:"table"`
	BatchSize     int            `yaml:"batch_size" mapstructure:"batch_size"`
	FlushInterval string         `yaml:"flush_interval" mapstructure:"flush_interval"`
}

type RotationConfig struct {
	MaxSize    int `yaml:"max_size" mapstructure:"max_size"`
	MaxAge     int `yaml:"max_age" mapstructure:"max_age"`
	MaxBackups int `yaml:"max_backups" mapstructure:"max_backups"`
}

type ServerConfig struct {
	HTTPPort int `yaml:"http_port" mapstructure:"http_port"`
	GRPCPort int `yaml:"grpc_port" mapstructure:"grpc_port"`
}

type WorkerConfig struct {
	MaxConcurrent int `yaml:"max_concurrent" mapstructure:"max_concurrent"`
	QueueSize     int `yaml:"queue_size" mapstructure:"queue_size"`
}

type SchedulerConfig struct {
	CheckInterval string `yaml:"check_interval" mapstructure:"check_interval"`
	Type          string `yaml:"type" mapstructure:"type"`
}

type MetricsConfig struct {
	Type string `yaml:"type" mapstructure:"type"`
	Port int    `yaml:"port" mapstructure:"port"`
}

func (c *Config) GetDSN() string {
	switch c.Database.Type {
	case "sqlite":
		return c.Database.DSN
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Database.Host, c.Database.Port, c.Database.User,
			c.Database.Password, c.Database.DBName)
	default:
		return ""
	}
}

func (c *Config) GetCheckInterval() time.Duration {
	if c.Scheduler.CheckInterval == "" {
		return 10 * time.Second
	}
	d, err := time.ParseDuration(c.Scheduler.CheckInterval)
	if err != nil {
		return 10 * time.Second
	}
	return d
}
