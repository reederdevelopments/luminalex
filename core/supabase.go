package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SupabaseClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSupabaseClient(url, key string) *SupabaseClient {
	return &SupabaseClient{
		baseURL: url,
		apiKey:  key,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *SupabaseClient) FetchUpdatedAfter(ctx context.Context, category string, after time.Time) ([]ContactRecord, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/%s?select=id,fields,updated_at,deleted&updated_at=gt.%s",
		c.baseURL, category, after.UTC().Format(time.RFC3339Nano))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create req: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase returned status %d", resp.StatusCode)
	}

	var remoteRows []struct {
		ID        string    `json:"id"`
		Fields    []string  `json:"fields"`
		UpdatedAt time.Time `json:"updated_at"`
		Deleted   bool      `json:"deleted"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&remoteRows); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	records := make([]ContactRecord, 0, len(remoteRows))
	for _, row := range remoteRows {
		records = append(records, ContactRecord{
			ID:        row.ID,
			Category:  category,
			Fields:    row.Fields,
			UpdatedAt: row.UpdatedAt,
			Deleted:   row.Deleted,
			Synced:    true,
		})
	}

	return records, nil
}

func (c *SupabaseClient) UpsertRecord(ctx context.Context, record ContactRecord) error {
	endpoint := fmt.Sprintf("%s/rest/v1/%s", c.baseURL, record.Category)

	payload := map[string]interface{}{
		"id":         record.ID,
		"fields":     record.Fields,
		"updated_at": record.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"deleted":    record.Deleted,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("create req: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase upsert failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *SupabaseClient) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}
