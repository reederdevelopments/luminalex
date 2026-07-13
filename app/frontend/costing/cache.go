package costing

import (
	"context"
	"log"
	"sync"
	"time"
)

type CacheEntry struct {
	TableData []ProjectCost
	Totals    ProjectCost
	ChartData []ChartPoint
	ExpiresAt time.Time
}

type CostCache struct {
	mu    sync.RWMutex
	store map[string]CacheEntry
	l     *log.Logger
}

func NewCostCache(l *log.Logger) *CostCache {
	return &CostCache{
		store: make(map[string]CacheEntry),
		l:     l,
	}
}

func (c *CostCache) Get(key string) (CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, found := c.store[key]
	if found && time.Now().Before(entry.ExpiresAt) {
		return entry, true
	}
	return CacheEntry{}, false
}

func (c *CostCache) Set(key string, table []ProjectCost, totals ProjectCost, chart []ChartPoint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = CacheEntry{
		TableData: table,
		Totals:    totals,
		ChartData: chart,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func (m module) preloadCacheWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	m.runComprehensivePreload()

	for range ticker.C {
		m.runComprehensivePreload()
	}
}

func (m module) runComprehensivePreload() {
	now := time.Now()
	projects := []string{"", "df-ps-south-africa", "df-ps-zambia", "df-ps-kenya", "df-ps-uganda", "df-ps-tanzania"}

	m.l.Printf("CACHE WORKER: Starting comprehensive 3-month preload sequence...")

	for i := 0; i < 3; i++ {
		targetDate := now.AddDate(0, -i, 0)
		firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		startStr := firstOfMonth.Format("2006-01-02")
		endStr := lastOfMonth.Format("2006-01-02")

		for _, proj := range projects {
			_, _, _, _ = m.fetchGCPBilling(context.Background(), startStr, endStr, proj, "daily")
		}
	}

	firstOfThreeMonthsAgo := time.Date(now.Year(), now.Month()-2, 1, 0, 0, 0, 0, now.Location())
	lastOfCurrentMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())

	for _, proj := range projects {
		_, _, _, _ = m.fetchGCPBilling(context.Background(), firstOfThreeMonthsAgo.Format("2006-01-02"), lastOfCurrentMonth.Format("2006-01-02"), proj, "daily")
	}

	m.l.Printf("CACHE WORKER: Preload complete. Dashboard is hot.")
}
