package transport

import (
	"context"
	"fmt"
)

type NATSBinding struct{}

func NewNATSBinding(servers []string) *NATSBinding {
	return &NATSBinding{}
}

func (b *NATSBinding) Connect(ctx context.Context) error {
	return fmt.Errorf("NATS binding not implemented yet")
}

func (b *NATSBinding) Close() error {
	return nil
}

func (b *NATSBinding) GetScheduler() SchedulerClient {
	return nil
}

func (b *NATSBinding) GetWorker() WorkerClient {
	return nil
}

func (b *NATSBinding) GetTaskManager() TaskManagerClient {
	return nil
}

func (b *NATSBinding) GetMetricAgent() MetricAgentClient {
	return nil
}

func (b *NATSBinding) RegisterScheduler(service SchedulerService) error {
	return fmt.Errorf("NATS binding does not support service registration")
}

func (b *NATSBinding) RegisterWorker(service WorkerService) error {
	return fmt.Errorf("NATS binding does not support service registration")
}

func (b *NATSBinding) RegisterTaskManager(service TaskManagerService) error {
	return fmt.Errorf("NATS binding does not support service registration")
}

func (b *NATSBinding) RegisterMetricAgent(service MetricAgentService) error {
	return fmt.Errorf("NATS binding does not support service registration")
}

func (b *NATSBinding) Health(ctx context.Context) error {
	return nil
}

func (b *NATSBinding) GetType() BindingType {
	return NATSType
}
