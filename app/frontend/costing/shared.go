package costing

import (
	"fmt"
	"strings"
	"time"
)

type ProjectCost struct {
	Project         string
	Total           string
	TotalDelta      string
	Datastream      string
	DatastreamDelta string
	DS_Size         string
	GCS             string
	GCSDelta        string
	GCS_Size        string
	Functions       string
	FunctionsDelta  string
	Func_Size       string
}

type ProjectCostRaw struct {
	Total  float64
	DS     float64
	DSGB   float64
	GCS    float64
	GCSGB  float64
	Func   float64
	FuncGB float64
}

type MonthData struct {
	MonthName string
	Cost      string
	Delta     string
}

type ResourceCostEx struct {
	Service     string
	SKU         string
	Description string
	TotalCost   string
	TotalNum    float64
	Months      []MonthData
}

type ProjectDetailsData struct {
	MonthHeaders []string
	Details      []ResourceCostEx
}

type ChartPoint struct {
	Date    string  `json:"date"`
	Project string  `json:"project"`
	Service string  `json:"service"`
	Cost    float64 `json:"cost"`
}

type CostingDashboardData struct {
	StartDate string
	EndDate   string
	Labels    []string          // Used to populate dynamic dropdowns
	LabelMap  map[string]string // Used for injecting human readable mappings
}

func formatSizeGB(gbAmount float64) string {
	if gbAmount < 0.01 && gbAmount > 0 {
		return "< 0.01 GB"
	} else if gbAmount >= 1024 {
		return fmt.Sprintf("%.2f TB", gbAmount/1024)
	}
	return fmt.Sprintf("%.2f GB", gbAmount)
}

func calcDelta(curr, prev float64) string {
	if prev == 0 {
		if curr > 0 {
			return "▲ 100%"
		}
		return "-"
	}
	delta := ((curr - prev) / prev) * 100
	if delta > 0 {
		return fmt.Sprintf("▲ %.1f%%", delta)
	} else if delta < 0 {
		return fmt.Sprintf("▼ %.1f%%", -delta)
	}
	return "0%"
}

func formatMonthHeader(ym string) string {
	t, err := time.Parse("2006-01", ym)
	if err != nil {
		return ym
	}
	return t.Format("Jan 2006")
}

func getSKUDescription(sku string) string {
	lower := strings.ToLower(sku)
	if strings.Contains(lower, "class a") {
		return "Charges for API operations that change state (insert/update)."
	} else if strings.Contains(lower, "class b") {
		return "Charges for API operations that read state (get/list)."
	} else if strings.Contains(lower, "early termination") {
		return "Fees for deleting/modifying lower-tier objects before minimum duration."
	} else if strings.Contains(lower, "storage") {
		return "Physical data at rest."
	} else if strings.Contains(lower, "compute") || strings.Contains(lower, "invocation") || strings.Contains(lower, "allocation") {
		return "Charges for CPU, memory, and trigger executions."
	}
	return "Standard GCP billing SKU or Job Name."
}
