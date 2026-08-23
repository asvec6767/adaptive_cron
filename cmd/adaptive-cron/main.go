package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"adaptive-cron/internal/app"
	"adaptive-cron/internal/config"
	"adaptive-cron/internal/features"
	"adaptive-cron/pkg/logger"
)

func main() {
	var (
		profile      string
		configFile   string
		showProfiles bool
		enableFeats  string
		disableFeats string
		version      bool
	)

	flag.StringVar(&profile, "profile", "lite", "Profile: lite, standard, enterprise")
	flag.StringVar(&configFile, "config", "", "Config file path")
	flag.BoolVar(&showProfiles, "list-profiles", false, "List available profiles")
	flag.StringVar(&enableFeats, "enable", "", "Enable additional features (comma-separated)")
	flag.StringVar(&disableFeats, "disable", "", "Disable features (comma-separated)")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.Parse()

	if version {
		fmt.Println("adaptive-cron version 1.0.0")
		return
	}

	if showProfiles {
		showAvailableProfiles()
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(profile, configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(logger.Config{
		Level:   cfg.Logging.Level,
		Format:  cfg.Logging.Format,
		Service: "adaptive-cron",
		File:    cfg.Logging.File,
		Rotation: logger.RotationConfig{
			MaxSize:    cfg.Logging.Rotation.MaxSize,
			MaxAge:     cfg.Logging.Rotation.MaxAge,
			MaxBackups: cfg.Logging.Rotation.MaxBackups,
		},
	})

	// Initialize features from config
	features.InitFromConfig(cfg.Features)

	// Apply additional feature flags
	if enableFeats != "" {
		for _, f := range strings.Split(enableFeats, ",") {
			if err := features.Enable(strings.TrimSpace(f)); err != nil {
				logger.Warnf("Failed to enable feature %s: %v", f, err)
			}
		}
	}
	if disableFeats != "" {
		for _, f := range strings.Split(disableFeats, ",") {
			if err := features.Disable(strings.TrimSpace(f)); err != nil {
				logger.Warnf("Failed to disable feature %s: %v", f, err)
			}
		}
	}

	logger.Infof("Starting Adaptive Cron with profile: %s", profile)
	logger.Infof("Enabled features: %v", features.GetEnabled())

	// Create application
	application, err := app.NewApp(cfg)
	if err != nil {
		logger.Fatalf("Failed to create app: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Run application
	if err := application.Run(ctx); err != nil {
		logger.Fatalf("Failed to run app: %v", err)
	}

	logger.Info("Application stopped")
}

func showAvailableProfiles() {
	profiles := config.AvailableProfiles
	fmt.Println("Available profiles:")
	fmt.Println()
	for _, p := range profiles {
		fmt.Printf("  %-12s - %s\n", p.Name, p.Description)
		fmt.Printf("               Features: %v\n", p.Features)
		fmt.Println()
	}
	fmt.Println("Usage:")
	fmt.Println("  adaptive-cron --profile lite")
	fmt.Println("  adaptive-cron --profile standard --config configs/standard.yaml")
	fmt.Println("  adaptive-cron --profile enterprise --config configs/enterprise.yaml")
	fmt.Println()
	fmt.Println("Feature flags:")
	fmt.Println("  --enable  feature1,feature2  Enable additional features")
	fmt.Println("  --disable feature1,feature2  Disable features")
}
