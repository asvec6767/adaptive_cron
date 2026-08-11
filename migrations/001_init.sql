-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    command TEXT NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    is_critical BOOLEAN DEFAULT FALSE,
    max_load_score FLOAT DEFAULT 0.7,
    max_cpu_percent FLOAT DEFAULT 80,
    max_memory_mb BIGINT DEFAULT 1024,
    timeout_seconds INT DEFAULT 300,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Job logs table
CREATE TABLE IF NOT EXISTS job_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    task_name VARCHAR(255),
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    exit_code INT,
    output TEXT,
    error TEXT,
    metrics_at_start JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Tasks indexes
CREATE INDEX idx_tasks_name ON tasks(name);
CREATE INDEX idx_tasks_is_active ON tasks(is_active);
CREATE INDEX idx_tasks_is_active_created_at ON tasks(is_active, created_at DESC);
CREATE INDEX idx_tasks_is_critical ON tasks(is_critical) WHERE is_critical = true;
CREATE INDEX idx_tasks_active_critical ON tasks(is_active, is_critical, max_load_score);
CREATE INDEX idx_tasks_schedule ON tasks(schedule) WHERE is_active = true;

-- Create Job Logs indexes
CREATE INDEX idx_job_logs_task_id ON job_logs(task_id);
CREATE INDEX idx_job_logs_start_time ON job_logs(start_time);
CREATE INDEX idx_job_logs_task_id_start_time ON job_logs(task_id, start_time DESC);
CREATE INDEX idx_job_logs_start_end_time ON job_logs(start_time, end_time);
CREATE INDEX idx_job_logs_task_name ON job_logs(task_name);
CREATE INDEX idx_job_logs_metrics_gin ON job_logs USING GIN (metrics_at_start);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add trigger to tasks table
CREATE TRIGGER update_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample tasks
INSERT INTO tasks (name, command, schedule, is_critical, max_load_score, max_cpu_percent, max_memory_mb) VALUES
('sample-backup', 'echo "Running backup..." && sleep 5', '0 2 * * *', false, 0.7, 80, 1024),
('sample-health', 'echo "Health check OK"', '*/5 * * * *', true, 0.9, 90, 512),
('sample-cleanup', 'echo "Cleaning up logs..."', '0 0 * * 0', false, 0.5, 50, 2048)
ON CONFLICT (name) DO NOTHING;