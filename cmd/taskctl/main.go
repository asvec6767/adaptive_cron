package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"adaptive-cron/internal/common/models"
	"adaptive-cron/internal/common/restclient"
	"adaptive-cron/pkg/logger"
)

var (
	apiURL  string
	client  *restclient.Client
	rootCmd *cobra.Command
)

type TaskInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Command    string    `json:"command"`
	Schedule   string    `json:"schedule"`
	IsCritical bool      `json:"is_critical"`
	IsActive   bool      `json:"is_active"`
	MaxCPU     float64   `json:"max_cpu_percent"`
	MaxMemory  int64     `json:"max_memory_mb"`
	CreatedAt  time.Time `json:"created_at"`
}

func init() {
	viper.SetDefault("api_url", "http://localhost:8080")
	viper.AutomaticEnv()

	rootCmd = &cobra.Command{
		Use:   "taskctl",
		Short: "TaskCtl - Adaptive Cron CLI",
		Long:  "A command-line interface for managing adaptive cron tasks",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			apiURL = viper.GetString("api_url")
			client = restclient.NewClient(apiURL, 30*time.Second)
			logger.Init("info", "text")
		},
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "api", viper.GetString("api_url"), "API Gateway URL")

	// Add commands
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(getCmd())
	rootCmd.AddCommand(updateCmd())
	rootCmd.AddCommand(deleteCmd())
	rootCmd.AddCommand(toggleCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(queueCmd())
	rootCmd.AddCommand(logsCmd())
	rootCmd.AddCommand(metricsCmd())
	rootCmd.AddCommand(statusCmd())
}

func addCmd() *cobra.Command {
	var (
		name      string
		command   string
		schedule  string
		critical  bool
		maxLoad   float64
		maxCPU    float64
		maxMemory int64
		timeout   int
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new task",
		RunE: func(cmd *cobra.Command, args []string) error {
			task := models.CreateTaskRequest{
				Name:           name,
				Command:        command,
				Schedule:       schedule,
				IsCritical:     critical,
				MaxLoadScore:   maxLoad,
				MaxCPUPercent:  maxCPU,
				MaxMemoryMB:    maxMemory,
				TimeoutSeconds: timeout,
			}

			var result models.Task
			ctx := cmd.Context()
			if err := client.Post(ctx, "/api/v1/tasks", task, &result); err != nil {
				return fmt.Errorf("failed to add task: %w", err)
			}

			fmt.Printf("✓ Task added successfully\n")
			fmt.Printf("  ID: %s\n", result.ID)
			fmt.Printf("  Name: %s\n", result.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Task name (required)")
	cmd.Flags().StringVarP(&command, "cmd", "c", "", "Command to execute (required)")
	cmd.Flags().StringVarP(&schedule, "schedule", "s", "", "Cron schedule (required)")
	cmd.Flags().BoolVar(&critical, "critical", false, "Mark as critical task")
	cmd.Flags().Float64Var(&maxLoad, "max-load", 0.7, "Maximum load score (0-1)")
	cmd.Flags().Float64Var(&maxCPU, "max-cpu", 80, "Maximum CPU percentage")
	cmd.Flags().Int64Var(&maxMemory, "max-memory", 1024, "Maximum memory in MB")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("cmd")
	cmd.MarkFlagRequired("schedule")

	return cmd
}

func listCmd() *cobra.Command {
	var active bool
	var critical bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var tasks []models.Task
			path := "/api/v1/tasks"

			if active {
				path += "?active=true"
			} else if critical {
				path += "?critical=true"
			}

			ctx := cmd.Context()
			if err := client.Get(ctx, path, &tasks); err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Println("No tasks found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSCHEDULE\tCRITICAL\tACTIVE\tMAX CPU\tMAX MEM")
			fmt.Fprintln(w, "--\t----\t--------\t--------\t------\t-------\t-------")

			for _, task := range tasks {
				criticalFlag := "❌"
				if task.IsCritical {
					criticalFlag = "✓"
				}
				activeFlag := "❌"
				if task.IsActive {
					activeFlag = "✓"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.0f%%\t%dMB\n",
					task.ID[:8], task.Name, task.Schedule, criticalFlag, activeFlag,
					task.MaxCPUPercent, task.MaxMemoryMB)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolVar(&active, "active", false, "Show only active tasks")
	cmd.Flags().BoolVar(&critical, "critical", false, "Show only critical tasks")

	return cmd
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <task-id>",
		Short: "Get task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			var task models.Task
			ctx := cmd.Context()
			if err := client.Get(ctx, "/api/v1/tasks/"+taskID, &task); err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}

			fmt.Printf("ID:          %s\n", task.ID)
			fmt.Printf("Name:        %s\n", task.Name)
			fmt.Printf("Command:     %s\n", task.Command)
			fmt.Printf("Schedule:    %s\n", task.Schedule)
			fmt.Printf("Critical:    %v\n", task.IsCritical)
			fmt.Printf("Active:      %v\n", task.IsActive)
			fmt.Printf("Max Load:    %.2f\n", task.MaxLoadScore)
			fmt.Printf("Max CPU:     %.0f%%\n", task.MaxCPUPercent)
			fmt.Printf("Max Memory:  %dMB\n", task.MaxMemoryMB)
			fmt.Printf("Timeout:     %ds\n", task.TimeoutSeconds)
			fmt.Printf("Created:     %s\n", task.CreatedAt.Format(time.RFC3339))

			return nil
		},
	}
}

func updateCmd() *cobra.Command {
	var (
		name      string
		command   string
		schedule  string
		critical  bool
		maxLoad   float64
		maxCPU    float64
		maxMemory int64
		timeout   int
		active    bool
	)

	cmd := &cobra.Command{
		Use:   "edit <task-id>",
		Short: "Update a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			updates := models.UpdateTaskRequest{}

			if cmd.Flags().Changed("name") {
				updates.Name = &name
			}
			if cmd.Flags().Changed("cmd") {
				updates.Command = &command
			}
			if cmd.Flags().Changed("schedule") {
				updates.Schedule = &schedule
			}
			if cmd.Flags().Changed("critical") {
				updates.IsCritical = &critical
			}
			if cmd.Flags().Changed("max-load") {
				updates.MaxLoadScore = &maxLoad
			}
			if cmd.Flags().Changed("max-cpu") {
				updates.MaxCPUPercent = &maxCPU
			}
			if cmd.Flags().Changed("max-memory") {
				updates.MaxMemoryMB = &maxMemory
			}
			if cmd.Flags().Changed("timeout") {
				updates.TimeoutSeconds = &timeout
			}
			if cmd.Flags().Changed("active") {
				updates.IsActive = &active
			}

			var result models.Task
			ctx := cmd.Context()
			if err := client.Put(ctx, "/api/v1/tasks/"+taskID, updates, &result); err != nil {
				return fmt.Errorf("failed to update task: %w", err)
			}

			fmt.Printf("✓ Task updated successfully\n")
			fmt.Printf("  Name: %s\n", result.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Task name")
	cmd.Flags().StringVar(&command, "cmd", "", "Command to execute")
	cmd.Flags().StringVar(&schedule, "schedule", "", "Cron schedule")
	cmd.Flags().BoolVar(&critical, "critical", false, "Mark as critical task")
	cmd.Flags().Float64Var(&maxLoad, "max-load", 0, "Maximum load score (0-1)")
	cmd.Flags().Float64Var(&maxCPU, "max-cpu", 0, "Maximum CPU percentage")
	cmd.Flags().Int64Var(&maxMemory, "max-memory", 0, "Maximum memory in MB")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout in seconds")
	cmd.Flags().BoolVar(&active, "active", false, "Enable/disable task")

	return cmd
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <task-id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			ctx := cmd.Context()
			if err := client.Delete(ctx, "/api/v1/tasks/"+taskID); err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}

			fmt.Printf("✓ Task deleted successfully\n")
			return nil
		},
	}
}

func toggleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "toggle <task-id>",
		Short: "Toggle task active status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			var task models.Task
			ctx := cmd.Context()
			if err := client.Post(ctx, "/api/v1/tasks/"+taskID+"/toggle", nil, &task); err != nil {
				return fmt.Errorf("failed to toggle task: %w", err)
			}

			status := "disabled"
			if task.IsActive {
				status = "enabled"
			}
			fmt.Printf("✓ Task %s\n", status)
			return nil
		},
	}
}

func runCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "run <task-id>",
		Short: "Force run a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			path := "/api/v1/scheduler/trigger/" + taskID
			if force {
				path += "?force=true"
			}

			var result map[string]interface{}
			ctx := cmd.Context()
			if err := client.Post(ctx, path, nil, &result); err != nil {
				return fmt.Errorf("failed to trigger task: %w", err)
			}

			fmt.Printf("✓ Task triggered\n")
			for k, v := range result {
				fmt.Printf("  %s: %v\n", k, v)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force run even if resources are constrained")

	return cmd
}

func queueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage delayed queue",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List delayed tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var result map[string]interface{}
			ctx := cmd.Context()
			if err := client.Get(ctx, "/api/v1/scheduler/queue", &result); err != nil {
				return fmt.Errorf("failed to get queue: %w", err)
			}

			tasks, ok := result["tasks"].([]interface{})
			if !ok || len(tasks) == 0 {
				fmt.Println("No delayed tasks")
				return nil
			}

			fmt.Printf("Delayed tasks: %d\n", len(tasks))
			for _, t := range tasks {
				fmt.Printf("  - %v\n", t)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "clear <task-id>",
		Short: "Remove task from delayed queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			ctx := cmd.Context()
			if err := client.Delete(ctx, "/api/v1/scheduler/queue/"+taskID); err != nil {
				return fmt.Errorf("failed to remove from queue: %w", err)
			}

			fmt.Printf("✓ Task removed from delayed queue\n")
			return nil
		},
	})

	return cmd
}

func logsCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "logs <task-id>",
		Short: "Show task execution logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// This would need a logs endpoint, for now placeholder
			fmt.Printf("Logs for task %s (limit: %d)\n", taskID, limit)
			fmt.Println("Logs endpoint coming soon...")
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Number of log entries to show")

	return cmd
}

func metricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "Show system metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			var metrics map[string]interface{}
			ctx := cmd.Context()
			if err := client.Get(ctx, "/api/v1/metrics/current", &metrics); err != nil {
				return fmt.Errorf("failed to get metrics: %w", err)
			}

			fmt.Printf("System Metrics:\n")
			fmt.Printf("  CPU:        %.2f%%\n", metrics["cpu_percent"])
			fmt.Printf("  Memory:     %.2f%%\n", metrics["memory_percent"])
			fmt.Printf("  Load Avg:   %.2f, %.2f, %.2f\n",
				metrics["load_avg1"], metrics["load_avg5"], metrics["load_avg15"])

			if score, ok := metrics["load_score"]; ok {
				fmt.Printf("  Load Score: %.2f\n", score)
			}

			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			var health map[string]string
			ctx := cmd.Context()
			if err := client.Get(ctx, "/health", &health); err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			fmt.Printf("System Status:\n")
			for service, status := range health {
				fmt.Printf("  %-15s: %s\n", service, status)
			}
			return nil
		},
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
