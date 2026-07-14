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
	Labels    []string

	// Structs for the specialized tabs
	TPTable TPTableData
	TPChart []TPChartPoint
	DFTable DFTableData

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
	entry := c.store[key]
	entry.TableData = table
	entry.Totals = totals
	entry.ChartData = chart
	entry.ExpiresAt = time.Now().Add(90 * 24 * time.Hour) // 90 Days
	c.store[key] = entry
}

func (c *CostCache) SetTP(key string, table TPTableData, chart []TPChartPoint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := c.store[key]
	entry.TPTable = table
	entry.TPChart = chart
	entry.ExpiresAt = time.Now().Add(90 * 24 * time.Hour)
	c.store[key] = entry
}

func (c *CostCache) SetDF(key string, table DFTableData, chart []ChartPoint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := c.store[key]
	entry.DFTable = table
	entry.ChartData = chart
	entry.ExpiresAt = time.Now().Add(90 * 24 * time.Hour)
	c.store[key] = entry
}

func (c *CostCache) SetLabels(key string, labels []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := c.store[key]
	entry.Labels = labels
	entry.ExpiresAt = time.Now().Add(90 * 24 * time.Hour)
	c.store[key] = entry
}

func (m module) preloadCacheWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	sastZone := time.FixedZone("SAST", 2*3600)

	// Always run once on startup
	m.runComprehensivePreload()

	for range ticker.C {
		nowSAST := time.Now().In(sastZone)
		hour := nowSAST.Hour()
		if hour >= 6 && hour <= 19 {
			m.runComprehensivePreload()
		} else {
			m.l.Printf("CACHE WORKER: Skipping 1-hour preload. Current SAST hour (%d) is outside 06:00-19:00 window.", hour)
		}
	}
}

func (m module) preloadFastCacheWorker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	sastZone := time.FixedZone("SAST", 2*3600)

	// Fast cache runs once on startup too
	m.runFastPreload()

	for range ticker.C {
		nowSAST := time.Now().In(sastZone)
		hour := nowSAST.Hour()
		if hour >= 6 && hour <= 19 {
			m.runFastPreload()
		}
	}
}

func (m module) runFastPreload() {
	now := time.Now()
	startStr := now.AddDate(0, 0, -90).Format("2006-01-02")
	endStr := now.Format("2006-01-02")

	m.l.Printf("CACHE WORKER: Fast 10-min preload for 3rd Party (Last 90 Days)...")
	ctx := context.Background()
	_, _, _ = m.fetchThirdPartyBilling(ctx, startStr, endStr, "daily")
	_, _, _ = m.fetchThirdPartyBilling(ctx, startStr, endStr, "monthly")
}

func (m module) runComprehensivePreload() {
	now := time.Now()
	m.l.Printf("CACHE WORKER: Executing 90-day preload sequence for all tabs...")

	startStr := now.AddDate(0, 0, -90).Format("2006-01-02")
	endStr := now.Format("2006-01-02")
	ctx := context.Background()

	// Preload 3rd Party
	_, _, _ = m.fetchThirdPartyBilling(ctx, startStr, endStr, "daily")
	_, _, _ = m.fetchThirdPartyBilling(ctx, startStr, endStr, "monthly")

	// Preload Dataform
	_, _ = m.fetchDataformLabels(ctx)
	_, _, _ = m.fetchDataformBilling(ctx, startStr, endStr, "daily", "")
	_, _, _ = m.fetchDataformBilling(ctx, startStr, endStr, "monthly", "")

	// Preload Datastream
	projects := []string{"", "df-ps-south-africa", "df-ps-zambia", "df-ps-kenya", "df-ps-uganda", "df-ps-tanzania"}
	for _, proj := range projects {
		_, _, _, _ = m.fetchGCPBilling(ctx, startStr, endStr, proj, "daily")
		_, _, _, _ = m.fetchGCPBilling(ctx, startStr, endStr, proj, "monthly")
	}

	m.l.Printf("CACHE WORKER: Sequence complete. 90-day cache is hot.")
}
