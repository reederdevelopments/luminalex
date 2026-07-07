package dataform

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"ujuzi_reloaded/app/backend/auth"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	dataformapi "google.golang.org/api/dataform/v1beta1"
)

const (
	dataformProjectID  = "df-fs-insights"
	dataformLocation   = "europe-west9"
	dataformRepository = "DATAFORM"
	dataformPollSecs   = 10
	bqBatchesTable     = "DATAFORM_BATCHES"
	bqExecutionsTable  = "DATAFORM_EXECUTIONS"
)

var dataformBasePath = fmt.Sprintf("projects/%s/locations/%s/repositories/%s", dataformProjectID, dataformLocation, dataformRepository)

var (
	bqClients = make(map[string]*bigquery.Client)
	clientsMu sync.Mutex
)

func getBQClient(ctx context.Context, projectID string) (*bigquery.Client, error) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := bqClients[projectID]; ok {
		return client, nil
	}
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	bqClients[projectID] = client
	return client, nil
}

type Module struct {
	l            *log.Logger
	sessionStore auth.Store
}

func NewModule(l *log.Logger, sessionStore auth.Store) *Module {
	return &Module{
		l:            l,
		sessionStore: sessionStore,
	}
}

type LiveBatch struct {
	ID          string
	Workspace   string
	TargetTag   string
	RequestedBy string
	StartedAt   time.Time
	FinishedAt  time.Time
	Status      string
	Countries   []LiveCountry
}

type LiveCountry struct {
	Code      string
	SortOrder int
	Status    string
	Cycles    []LiveCycle
}

type LiveCycle struct {
	Cycle       string
	ExecutionID string
	ConsoleLink string
	Status      string
}

type BatchHistoryRow struct {
	ID          string    `bigquery:"ID"`
	Workspace   string    `bigquery:"WORKSPACE"`
	TargetTag   string    `bigquery:"TARGET_TAG"`
	RequestedBy string    `bigquery:"REQUESTED_BY"`
	StartedAt   time.Time `bigquery:"STARTED_AT"`
	FinishedAt  time.Time `bigquery:"FINISHED_AT"`
	Status      string    `bigquery:"STATUS"`
}

type ExecutionHistoryRow struct {
	BatchID     string `bigquery:"BATCH_ID"`
	CountryCode string `bigquery:"COUNTRY_CODE"`
	Cycle       string `bigquery:"CYCLE"`
	ExecutionID string `bigquery:"EXECUTION_ID"`
	ConsoleLink string `bigquery:"CONSOLE_LINK"`
	Status      string `bigquery:"STATUS"`
}

type DataformCache struct {
	mu         sync.RWMutex
	batch      *LiveBatch
	cancelFunc context.CancelFunc
}

var dfCache = &DataformCache{}

func (c *DataformCache) Get() (*LiveBatch, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.batch == nil {
		return nil, false
	}
	b := *c.batch
	return &b, true
}

func (c *DataformCache) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.batch != nil && c.batch.Status == "RUNNING"
}

func (c *DataformCache) Start(batch *LiveBatch, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batch = batch
	c.cancelFunc = cancel
}

func (c *DataformCache) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
}

func (c *DataformCache) Finish(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.batch != nil {
		c.batch.Status = status
		c.batch.FinishedAt = time.Now()
	}
	c.cancelFunc = nil
}

func (c *DataformCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batch = nil
}

func (m *Module) Loader(w http.ResponseWriter, r *http.Request) error {
	return dataformPage().Render(r.Context(), w)
}

func (m *Module) BatchRun(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	if dfCache.IsRunning() {
		w.Write([]byte(`<div class="p-4 bg-red-50 text-red-700 rounded font-bold text-sm">A batch is already running!</div>`))
		return nil
	}

	workspace := r.FormValue("workspace")
	targetTag := r.FormValue("target_tag")

	var liveCountries []LiveCountry
	orderIdx := 1
	for _, cc := range []string{"za", "ke", "ug", "tz", "zm"} {
		if r.FormValue(fmt.Sprintf("enable_%s", cc)) == "true" {
			start := r.FormValue(fmt.Sprintf("start_%s", cc))
			end := r.FormValue(fmt.Sprintf("end_%s", cc))

			cycles := []LiveCycle{}
			if start != "" && end != "" {
				cycles = append(cycles, LiveCycle{Cycle: start, Status: "AWAITING"})
				if start != end {
					cycles = append(cycles, LiveCycle{Cycle: end, Status: "AWAITING"})
				}
			}

			liveCountries = append(liveCountries, LiveCountry{
				Code:      cc,
				SortOrder: orderIdx,
				Status:    "AWAITING",
				Cycles:    cycles,
			})
			orderIdx++
		}
	}

	batch := &LiveBatch{
		ID:          uuid.New().String(),
		Workspace:   workspace,
		TargetTag:   targetTag,
		RequestedBy: auth.FromCtx(ctx).User.FirstName,
		StartedAt:   time.Now(),
		Status:      "RUNNING",
		Countries:   liveCountries,
	}

	batchCtx, cancel := context.WithCancel(context.Background())
	dfCache.Start(batch, cancel)

	go m.runDataformBatch(batchCtx, batch)

	return batchStatusPanel(batch, nil).Render(ctx, w)
}

func (m *Module) BatchCancel(w http.ResponseWriter, r *http.Request) error {
	dfCache.Cancel()
	batch, _ := dfCache.Get()
	return batchStatusPanel(batch, nil).Render(r.Context(), w)
}

func (m *Module) BatchStatus(w http.ResponseWriter, r *http.Request) error {
	batch, isRunning := dfCache.Get()

	var history []BatchHistoryRow
	if !isRunning {
		bqClient, err := getBQClient(r.Context(), "df-frontend")
		if err == nil {
			q := bqClient.Query("SELECT ID, WORKSPACE, TARGET_TAG, REQUESTED_BY, STARTED_AT, FINISHED_AT, STATUS FROM `df-frontend.UJUZI.DATAFORM_BATCHES` ORDER BY STARTED_AT DESC LIMIT 10")
			it, err := q.Read(r.Context())
			if err == nil {
				for {
					var row BatchHistoryRow
					if err := it.Next(&row); err != nil {
						break
					}
					history = append(history, row)
				}
			}
		}
	}

	return batchStatusPanel(batch, history).Render(r.Context(), w)
}

func (m *Module) runDataformBatch(ctx context.Context, batch *LiveBatch) {
	dfSvc, err := dataformapi.NewService(ctx)
	if err != nil {
		m.l.Printf("dataform: failed to create service: %v", err)
		dfCache.Finish("FAILED")
		return
	}

	workspacePath := fmt.Sprintf(
		"projects/%s/locations/%s/repositories/%s/workspaces/%s",
		dataformProjectID, dataformLocation, dataformRepository, batch.Workspace,
	)

	anyFailed := false

	for i, country := range batch.Countries {
		select {
		case <-ctx.Done():
			dfCache.Finish("CANCELLED")
			m.writeBatchToBQ(batch.ID)
			return
		default:
		}

		if country.Status == "SKIPPED" {
			continue
		}

		dfCache.mu.Lock()
		batch.Countries[i].Status = "RUNNING"
		dfCache.mu.Unlock()

		countryFailed := false
		for j, lc := range country.Cycles {
			select {
			case <-ctx.Done():
				dfCache.Finish("CANCELLED")
				m.writeBatchToBQ(batch.ID)
				return
			default:
			}

			dfCache.mu.Lock()
			batch.Countries[i].Cycles[j].Status = "RUNNING"
			dfCache.mu.Unlock()

			compilationName, err := m.compileWorkspace(ctx, dfSvc, workspacePath, country.Code, lc.Cycle)
			if err != nil {
				m.l.Printf("dataform: compile failed for %s/%s: %v", country.Code, lc.Cycle, err)
				dfCache.mu.Lock()
				batch.Countries[i].Cycles[j].Status = "FAILED"
				dfCache.mu.Unlock()
				countryFailed = true
				break
			}

			executionName, executionID, err := m.invokeWorkflow(ctx, dfSvc, compilationName, batch.TargetTag)
			if err != nil {
				m.l.Printf("dataform: invoke failed for %s/%s: %v", country.Code, lc.Cycle, err)
				dfCache.mu.Lock()
				batch.Countries[i].Cycles[j].Status = "FAILED"
				dfCache.mu.Unlock()
				countryFailed = true
				break
			}

			consoleLink := fmt.Sprintf(
				"https://console.cloud.google.com/bigquery/dataform/locations/%s/repositories/%s/workflows/%s?project=%s",
				dataformLocation, dataformRepository, executionID, dataformProjectID,
			)

			dfCache.mu.Lock()
			batch.Countries[i].Cycles[j].ExecutionID = executionID
			batch.Countries[i].Cycles[j].ConsoleLink = consoleLink
			dfCache.mu.Unlock()

			finalState, err := m.pollExecution(ctx, dfSvc, executionName)
			if err != nil || finalState != "SUCCEEDED" {
				m.l.Printf("dataform: %s/%s ended with state %s/err:%v", country.Code, lc.Cycle, finalState, err)
				dfCache.mu.Lock()
				batch.Countries[i].Cycles[j].Status = "FAILED"
				dfCache.mu.Unlock()
				countryFailed = true
				break
			}

			dfCache.mu.Lock()
			batch.Countries[i].Cycles[j].Status = "SUCCEEDED"
			dfCache.mu.Unlock()
		}

		dfCache.mu.Lock()
		if countryFailed {
			batch.Countries[i].Status = "FAILED"
			anyFailed = true
		} else {
			batch.Countries[i].Status = "COMPLETED"
		}
		dfCache.mu.Unlock()
	}

	finalStatus := "COMPLETED"
	if anyFailed {
		finalStatus = "FAILED"
	}
	dfCache.Finish(finalStatus)
	m.writeBatchToBQ(batch.ID)
}

func (m *Module) compileWorkspace(ctx context.Context, svc *dataformapi.Service, workspacePath, countryCode, cycle string) (string, error) {
	result, err := svc.Projects.Locations.Repositories.CompilationResults.
		Create(dataformBasePath, &dataformapi.CompilationResult{
			Workspace: workspacePath,
			CodeCompilationConfig: &dataformapi.CodeCompilationConfig{
				Vars: map[string]string{"countryCode": countryCode, "cycle": cycle},
			},
		}).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return result.Name, nil
}

func (m *Module) invokeWorkflow(ctx context.Context, svc *dataformapi.Service, compilationName, targetTag string) (executionName, executionID string, err error) {
	result, err := svc.Projects.Locations.Repositories.WorkflowInvocations.
		Create(dataformBasePath, &dataformapi.WorkflowInvocation{
			CompilationResult: compilationName,
			InvocationConfig: &dataformapi.InvocationConfig{
				IncludedTags:                   []string{targetTag},
				TransitiveDependenciesIncluded: true,
			},
		}).Context(ctx).Do()
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(result.Name, "/")
	id := parts[len(parts)-1]
	return result.Name, id, nil
}

func (m *Module) pollExecution(ctx context.Context, svc *dataformapi.Service, executionName string) (string, error) {
	ticker := time.NewTicker(dataformPollSecs * time.Second)
	defer ticker.Stop()

	consecutiveErrors := 0
	const maxErrors = 5

	for {
		select {
		case <-ctx.Done():
			return "CANCELLED", ctx.Err()
		case <-ticker.C:
			inv, err := svc.Projects.Locations.Repositories.WorkflowInvocations.Get(executionName).Context(ctx).Do()
			if err != nil {
				consecutiveErrors++
				m.l.Printf("WARN: dataform polling transient error (%d/%d) for %s: %v", consecutiveErrors, maxErrors, executionName, err)

				if consecutiveErrors >= maxErrors {
					return "", fmt.Errorf("polling failed after %d consecutive errors: %w", maxErrors, err)
				}
				continue // Skip the rest of this loop iteration and wait for the next tick
			}

			// Reset the error counter on a successful API response
			consecutiveErrors = 0

			switch inv.State {
			case "RUNNING", "CANCELING", "QUEUED":
				// keep polling
			default:
				return inv.State, nil
			}
		}
	}
}

func (m *Module) writeBatchToBQ(batchID string) {
	batch, ok := dfCache.Get()
	if !ok {
		return
	}
	bqClient, err := getBQClient(context.Background(), "df-frontend")
	if err != nil {
		m.l.Printf("Failed to get BQ client for history write: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	batchRow := &BatchHistoryRow{
		ID:          batch.ID,
		Workspace:   batch.Workspace,
		TargetTag:   batch.TargetTag,
		RequestedBy: batch.RequestedBy,
		StartedAt:   batch.StartedAt,
		FinishedAt:  batch.FinishedAt,
		Status:      batch.Status,
	}
	if err := bqClient.Dataset("UJUZI").Table(bqBatchesTable).Inserter().Put(ctx, batchRow); err != nil {
		m.l.Printf("dataform: failed to write batch to BQ: %v", err)
	}

	var execRows []*ExecutionHistoryRow
	for _, country := range batch.Countries {
		for _, cycle := range country.Cycles {
			if cycle.ExecutionID == "" {
				continue
			}
			execRows = append(execRows, &ExecutionHistoryRow{
				BatchID:     batch.ID,
				CountryCode: country.Code,
				Cycle:       cycle.Cycle,
				ExecutionID: cycle.ExecutionID,
				ConsoleLink: cycle.ConsoleLink,
				Status:      cycle.Status,
			})
		}
	}
	if len(execRows) > 0 {
		if err := bqClient.Dataset("UJUZI").Table(bqExecutionsTable).Inserter().Put(ctx, execRows); err != nil {
			m.l.Printf("dataform: failed to write executions to BQ: %v", err)
		}
	}
	dfCache.Clear()
}
