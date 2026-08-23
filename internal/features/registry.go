package features

import (
	"fmt"
	"strings"
	"sync"
)

type FeatureFlag struct {
	Name        string
	Description string
	Enabled     bool
	Category    string
}

type Registry struct {
	mu    sync.RWMutex
	flags map[string]*FeatureFlag
}

var globalRegistry = &Registry{
	flags: make(map[string]*FeatureFlag),
}

func Register(flag *FeatureFlag) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.flags[flag.Name] = flag
}

func RegisterAllFromList(names []string) {
	for _, name := range names {
		desc := FlagDescriptions[name]
		if desc == "" {
			desc = name
		}
		Register(&FeatureFlag{
			Name:        name,
			Description: desc,
			Enabled:     true,
			Category:    getCategory(name),
		})
	}
}

func getCategory(name string) string {
	parts := make(map[string]string)
	for _, flag := range AllFlags {
		parts[flag] = flag[:strings.Index(flag, ".")]
	}
	if cat, ok := parts[name]; ok {
		return cat
	}
	return "general"
}

func IsEnabled(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	if flag, ok := globalRegistry.flags[name]; ok {
		return flag.Enabled
	}
	return false
}

func Enable(name string) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if flag, ok := globalRegistry.flags[name]; ok {
		flag.Enabled = true
		return nil
	}
	return fmt.Errorf("feature %s not found", name)
}

func Disable(name string) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if flag, ok := globalRegistry.flags[name]; ok {
		flag.Enabled = false
		return nil
	}
	return fmt.Errorf("feature %s not found", name)
}

func GetAll() map[string]*FeatureFlag {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make(map[string]*FeatureFlag)
	for k, v := range globalRegistry.flags {
		result[k] = v
	}
	return result
}

func GetEnabled() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := []string{}
	for name, flag := range globalRegistry.flags {
		if flag.Enabled {
			result = append(result, name)
		}
	}
	return result
}

func InitFromConfig(features []string) {
	for _, name := range features {
		desc := FlagDescriptions[name]
		if desc == "" {
			desc = name
		}
		Register(&FeatureFlag{
			Name:        name,
			Description: desc,
			Enabled:     true,
			Category:    getCategory(name),
		})
	}
}
