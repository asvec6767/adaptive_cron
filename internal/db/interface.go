package db

import (
	"database/sql"
	"fmt"
)

type Database interface {
	Connect() error
	Close() error
	GetDB() *sql.DB
	Migrate() error
	Ping() error
}

type DatabaseFactory interface {
	Create(config Config) (Database, error)
}

// Регистрация фабрик
type DatabaseRegistry struct {
	factories map[string]DatabaseFactory
}

func NewDatabaseRegistry() *DatabaseRegistry {
	return &DatabaseRegistry{
		factories: make(map[string]DatabaseFactory),
	}
}

func (r *DatabaseRegistry) Register(name string, factory DatabaseFactory) {
	r.factories[name] = factory
}

func (r *DatabaseRegistry) Get(name string, config Config) (Database, error) {
	if factory, ok := r.factories[name]; ok {
		return factory.Create(config)
	}
	return nil, fmt.Errorf("database %s not registered", name)
}
