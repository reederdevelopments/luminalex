package monitoring

import (
	"log"
	"sync"
	"time"
)

type MonCache struct {
	mu           sync.RWMutex
	summaries    map[string][]JobSummary
	summariesExp map[string]time.Time
	logs         map[string][]LogEntry
	logsStart    time.Time
	logsEnd      time.Time
	l            *log.Logger
}

func NewMonCache(l *log.Logger) *MonCache {
	return &MonCache{
		summaries:    make(map[string][]JobSummary),
		summariesExp: make(map[string]time.Time),
		logs:         make(map[string][]LogEntry),
		l:            l,
	}
}

func (c *MonCache) GetSummaries(key string) ([]JobSummary, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	exp, ok := c.summariesExp[key]
	if ok && time.Now().Before(exp) {
		return c.summaries[key], true
	}
	return nil, false
}

func (c *MonCache) SetSummaries(key string, summaries []JobSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.summaries[key] = summaries
	c.summariesExp[key] = time.Now().Add(50 * time.Second)
}

func (c *MonCache) GetLogs(jobName string, start, end time.Time) ([]LogEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.logsStart.IsZero() || c.logsEnd.IsZero() {
		return nil, false
	}
	if start.Before(c.logsStart) || end.After(c.logsEnd) {
		return nil, false
	}
	entries, ok := c.logs[jobName]
	if !ok {
		return nil, false
	}
	var filtered []LogEntry
	for _, e := range entries {
		if !e.StartTime.Before(start) && !e.StartTime.After(end.Add(24*time.Hour)) {
			filtered = append(filtered, e)
		}
	}
	return filtered, true
}

func (c *MonCache) SetLogs(start, end time.Time, data map[string][]LogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logsStart = start
	c.logsEnd = end
	for k, v := range data {
		c.logs[k] = v
	}
}
