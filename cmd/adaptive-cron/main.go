package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"adaptive-cron/internal/app"
	"adaptive-cron/internal/config"
	"adaptive-cron/pkg/logger"
)

func main() {
	var (
		profile      string
		configFile   string
		showProfiles bool
		enableFeats  string
		disableFeats string
	)

	flag.StringVar(&profile, "profile", "lite", "Profile: lite, standard, enterprise")
	flag.StringVar(&configFile, "config", "", "Config file path")
	flag.BoolVar(&showProfiles, "list-profiles", false, "List available profiles")
	flag.StringVar(&enableFeats, "enable", "", "Enable additional features")
	flag.StringVar(&disableFeats, "disable", "", "Disable features")
	flag.Parse()

	if showProfiles {
		showAvailableProfiles()
		return
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(profile, configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация логгера
	logger.Init(logger.LoggerConfig{
		Level:   cfg.Logging.Level,
		Format:  cfg.Logging.Format,
		Service: "adaptive-cron",
	})

	// Инициализация feature flags
	initFeatures(profile, enableFeats, disableFeats)

	// Создание приложения
	application, err := app.NewApp(cfg)
	if err != nil {
		logger.Fatalf("Failed to create app: %v", err)
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Запуск
	if err := application.Run(ctx); err != nil {
		logger.Fatalf("Failed to run app: %v", err)
	}
}

func showAvailableProfiles() {
	fmt.Println("Available profiles:")
	fmt.Println()
	fmt.Println("  lite         - All services in one process (in-memory)")
	fmt.Println("  standard     - Services as separate processes (REST)")
	fmt.Println("  enterprise   - Distributed services (NATS/gRPC)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  adaptive-cron --profile lite")
	fmt.Println("  adaptive-cron --profile standard --config configs/standard.yaml")
	fmt.Println("  adaptive-cron --profile enterprise --config configs/enterprise.yaml")
}

func initFeatures(profile, enable, disable string) {
	// Регистрация features для профиля
	// ...
}
