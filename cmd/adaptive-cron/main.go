package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"adaptive-cron/internal/app"
	"adaptive-cron/internal/config"
	"adaptive-cron/internal/db"
	"adaptive-cron/internal/features"
	"adaptive-cron/internal/log"
	"adaptive-cron/internal/queue"
)

func main() {
	// Парсинг аргументов командной строки
	var (
		profile         string
		configFile      string
		showProfiles    bool
		enableFeatures  string
		disableFeatures string
	)

	flag.StringVar(&profile, "profile", "lite", "Profile to use (lite, standard, enterprise, cloud)")
	flag.StringVar(&configFile, "config", "config.yaml", "Config file path")
	flag.BoolVar(&showProfiles, "list-profiles", false, "List available profiles")
	flag.StringVar(&enableFeatures, "enable", "", "Enable additional features (comma-separated)")
	flag.StringVar(&disableFeatures, "disable", "", "Disable features (comma-separated)")
	flag.Parse()

	// Показать доступные профили
	if showProfiles {
		showAvailableProfiles()
		return
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(profile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация feature flags из конфига
	initFeaturesFromProfile(profile)

	// Применение дополнительных флагов
	if enableFeatures != "" {
		for _, f := range strings.Split(enableFeatures, ",") {
			features.Enable(strings.TrimSpace(f))
		}
	}
	if disableFeatures != "" {
		for _, f := range strings.Split(disableFeatures, ",") {
			features.Disable(strings.TrimSpace(f))
		}
	}

	// Создание приложения с учетом feature flags
	app, err := createApp(cfg)
	if err != nil {
		fmt.Printf("Failed to create app: %v\n", err)
		os.Exit(1)
	}

	// Запуск приложения
	if err := app.Run(); err != nil {
		fmt.Printf("Failed to run app: %v\n", err)
		os.Exit(1)
	}
}

func showAvailableProfiles() {
	fmt.Println("Available profiles:")
	fmt.Println()
	fmt.Println("  lite         - Lightweight version for small websites")
	fmt.Println("  standard     - Standard version with full logging")
	fmt.Println("  enterprise   - Enterprise version with distributed architecture")
	fmt.Println("  cloud        - Cloud-native version with full observability")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  adaptive-cron --profile lite")
	fmt.Println("  adaptive-cron --profile standard")
	fmt.Println("  adaptive-cron --profile enterprise --enable metrics.prometheus")
}

func initFeaturesFromProfile(profile string) {
	// Получение features для профиля
	featuresMap := features.ProfileFeatures[profile]
	for _, f := range featuresMap {
		features.Register(&features.FeatureFlag{
			Name:        f,
			Enabled:     true,
			Category:    getCategory(f),
			Description: getDescription(f),
		})
	}
}

func getCategory(feature string) string {
	parts := strings.Split(feature, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "general"
}

func getDescription(feature string) string {
	descriptions := map[string]string{
		"db.sqlite":             "Use SQLite as database",
		"db.postgres":           "Use PostgreSQL as database",
		"queue.memory":          "Use in-memory queue",
		"queue.redis":           "Use Redis as queue",
		"queue.nats":            "Use NATS as queue",
		"log.file":              "Write logs to file",
		"log.json":              "Use JSON log format",
		"log.clickhouse":        "Use ClickHouse for logs storage",
		"transport.rest":        "Enable REST API",
		"transport.grpc":        "Enable gRPC API",
		"metrics.prometheus":    "Enable Prometheus metrics",
		"scheduler.distributed": "Enable distributed scheduling",
	}
	if desc, ok := descriptions[feature]; ok {
		return desc
	}
	return feature
}

func createApp(cfg *config.Config) (*app.App, error) {
	// Создание компонентов на основе feature flags

	// 1. База данных
	dbFactory := db.NewDatabaseRegistry()
	dbFactory.Register("sqlite", &db.SQLiteFactory{})
	dbFactory.Register("postgres", &db.PostgresFactory{})

	var dbInstance db.Database
	if features.IsEnabled(features.FlagDBSQLite) {
		dbInstance, _ = dbFactory.Get("sqlite", cfg.Database)
	} else if features.IsEnabled(features.FlagDBPostgres) {
		dbInstance, _ = dbFactory.Get("postgres", cfg.Database)
	}

	// 2. Очередь
	queueFactory := queue.NewQueueRegistry()
	queueFactory.Register("memory", &queue.MemoryQueueFactory{})

	var queueInstance queue.Queue
	if features.IsEnabled(features.FlagQueueMemory) {
		queueInstance, _ = queueFactory.Get("memory", cfg.Queue)
	}

	// 3. Логирование
	var logger log.Logger
	if features.IsEnabled(features.FlagLogFile) {
		logger = log.NewFileLogger(cfg.Logging)
	}

	// 4. Приложение
	app := app.NewApp(
		app.WithDatabase(dbInstance),
		app.WithQueue(queueInstance),
		app.WithLogger(logger),
		app.WithConfig(cfg),
		app.WithFeatures(features.GetAll()),
	)

	return app, nil
}
