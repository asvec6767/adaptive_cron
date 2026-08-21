package features

import (
	"fmt"
	"sync"
)

// FeatureFlag определяет один feature
type FeatureFlag struct {
	Name        string
	Description string
	Enabled     bool
	Category    string
}

// FeatureRegistry хранит все feature flags
type FeatureRegistry struct {
	mu    sync.RWMutex
	flags map[string]*FeatureFlag
}

var globalRegistry = &FeatureRegistry{
	flags: make(map[string]*FeatureFlag),
}

// Регистрация feature
func Register(flag *FeatureFlag) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.flags[flag.Name] = flag
}

// Проверка включен ли feature
func IsEnabled(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	if flag, ok := globalRegistry.flags[name]; ok {
		return flag.Enabled
	}
	return false
}

// Включить feature
func Enable(name string) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if flag, ok := globalRegistry.flags[name]; ok {
		flag.Enabled = true
		return nil
	}
	return fmt.Errorf("feature %s not found", name)
}

// Отключить feature
func Disable(name string) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if flag, ok := globalRegistry.flags[name]; ok {
		flag.Enabled = false
		return nil
	}
	return fmt.Errorf("feature %s not found", name)
}

// Получить все features
func GetAll() map[string]*FeatureFlag {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make(map[string]*FeatureFlag)
	for k, v := range globalRegistry.flags {
		result[k] = v
	}
	return result
}
