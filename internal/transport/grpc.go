package transport

import (
	"context"
	"fmt"
)

type GRPCBinding struct{}

func NewGRPCBinding(address string) *GRPCBinding {
	return &GRPCBinding{}
}

func (b *GRPCBinding) Connect(ctx context.Context) error {
	return fmt.Errorf("gRPC binding not implemented yet")
}

func (b *GRPCBinding) Close() error {
	return nil
}

func (b *GRPCBinding) GetScheduler() SchedulerClient {
	return nil
}

func (b *GRPCBinding) GetWorker() WorkerClient {
	return nil
}

func (b *GRPCBinding) GetTaskManager() TaskManagerClient {
	return nil
}

func (b *GRPCBinding) GetMetricAgent() MetricAgentClient {
	return nil
}

func (b *GRPCBinding) RegisterScheduler(service SchedulerService) error {
	return fmt.Errorf("gRPC binding does not support service registration")
}

func (b *GRPCBinding) RegisterWorker(service WorkerService) error {
	return fmt.Errorf("gRPC binding does not support service registration")
}

func (b *GRPCBinding) RegisterTaskManager(service TaskManagerService) error {
	return fmt.Errorf("gRPC binding does not support service registration")
}

func (b *GRPCBinding) RegisterMetricAgent(service MetricAgentService) error {
	return fmt.Errorf("gRPC binding does not support service registration")
}

func (b *GRPCBinding) Health(ctx context.Context) error {
	return nil
}

func (b *GRPCBinding) GetType() BindingType {
	return GRPCBinding
}
