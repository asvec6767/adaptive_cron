package domain

import "errors"

var (
	ErrMetricsUnavailable  = errors.New("metrics not available")
	ErrCollectorNotRunning = errors.New("metric collector not running")
	ErrHistoryEmpty        = errors.New("metrics history is empty")
)
