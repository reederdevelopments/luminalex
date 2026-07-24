package monitoring

import (
	"fmt"
	"time"
)

type JobSummary struct {
	JobName     string
	Status      string
	LastRunTime time.Time
	TotalRows   int64
	TotalBytes  int64
	TotalCost   float64
}

type LogEntry struct {
	JobName         string
	StartTime       time.Time
	EndTime         time.Time
	DurationSeconds float64
	Status          string
	TotalRows       int64
	TotalBytes      int64
	MetricsJSON     string
	ErrorMessage    string
	StackTrace      string
}

type MetricDetail struct {
	Rows  int64 `json:"rows"`
	Bytes int64 `json:"bytes"`
}

type ParsedMetrics map[string]MetricDetail

type MonitoringDashboardData struct {
	StartDate string
	EndDate   string
}

func formatBytes(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", b/float64(div), "KMGTPE"[exp])
}
