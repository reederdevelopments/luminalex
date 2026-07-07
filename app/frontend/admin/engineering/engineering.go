package engineering

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"ujuzi_reloaded/app/backend/auth"

	"cloud.google.com/go/datastore"
)

type CoreReload struct {
	EncodedKey               string `datastore:"-"`
	Name                     string `datastore:"name"`
	Database                 string `datastore:"database"`
	PreviousTableRefreshSAST string `datastore:"previous_table_refresh_sast"`
	NextTableRefreshSAST     string `datastore:"next_table_refresh_sast"`
	CronSchedule             string `datastore:"cron_schedule"`
	HasGcsData               bool   `datastore:"has_gcs_data"`
	RefreshInProgress        bool   `datastore:"refresh_in_progress"`
	ExecutionID              string `datastore:"execution_id,noindex"`
}

type JobConfig struct {
	EncodedKey               string
	Name                     string
	Type                     string
	Active                   bool
	Schedule                 string
	Link                     string
	Repository               string
	UniqueTag                string
	CountryAgnostic          bool
	CronDefault              string
	CountriesOverride        string // JSON
	CronOverride             string // JSON
	UpdateSourceTablesConfig string // JSON
	MailingList              string // JSON
	Properties               map[string]interface{}
}

var (
	dsClients = make(map[string]*datastore.Client)
	clientsMu sync.Mutex
)

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

func getProjectForCountry(cc string) string {
	cc = strings.ToLower(cc)
	envKey := fmt.Sprintf("GOOGLE_%s_PROJECT_ID", strings.ToUpper(cc))
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	names := map[string]string{
		"za": "south-africa", "ke": "kenya", "ug": "uganda", "tz": "tanzania", "zm": "zambia",
	}
	return fmt.Sprintf("df-ps-%s", names[cc])
}

func getDatastoreClient(ctx context.Context, projectID string) (*datastore.Client, error) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := dsClients[projectID]; ok {
		return client, nil
	}
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dsClients[projectID] = client
	return client, nil
}

func (m *Module) AdminEngineeringLoader(w http.ResponseWriter, r *http.Request) error {
	return engineeringPage().Render(r.Context(), w)
}
