package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	var endpoint string

	if after.IsZero() {
		endpoint = fmt.Sprintf("%s/rest/v1/%s?select=id,fields,updated_at,deleted", c.baseURL, category)
	} else {
		endpoint = fmt.Sprintf("%s/rest/v1/%s?select=id,fields,updated_at,deleted&updated_at=gt.%s",
			c.baseURL, category, after.UTC().Format("2006-01-02T15:04:05.999999Z"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var remoteRows []struct {
		ID        string          `json:"id"`
		Fields    json.RawMessage `json:"fields"`
		UpdatedAt time.Time       `json:"updated_at"`
		Deleted   bool            `json:"deleted"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&remoteRows); err != nil {
		return nil, err
	}

	records := make([]ContactRecord, 0, len(remoteRows))
	for _, row := range remoteRows {
		var fields []string

		if len(row.Fields) > 0 {
			if row.Fields[0] == '"' {
				var unescaped string
				if err := json.Unmarshal(row.Fields, &unescaped); err == nil {
					_ = json.Unmarshal([]byte(unescaped), &fields)
				}
			} else {
				_ = json.Unmarshal(row.Fields, &fields)
			}
		}

		if fields == nil {
			fields = []string{}
		}

		records = append(records, ContactRecord{
			ID:        row.ID,
			Category:  category,
			Fields:    fields,
			UpdatedAt: row.UpdatedAt,
			Deleted:   row.Deleted,
			Synced:    true,
		})
	}

	return records, nil
}

func (c *SupabaseClient) FetchClientsAfter(ctx context.Context, after time.Time) ([]Client, error) {
	var endpoint string
	if after.IsZero() {
		endpoint = fmt.Sprintf("%s/rest/v1/clients?select=*", c.baseURL)
	} else {
		endpoint = fmt.Sprintf("%s/rest/v1/clients?select=*&updated_at=gt.%s",
			c.baseURL, after.UTC().Format("2006-01-02T15:04:05.999999Z"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var clients []Client
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func (c *SupabaseClient) FetchAddresses(ctx context.Context, clientIDs []string) ([]ClientAddress, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/client_addresses?select=*", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var addresses []ClientAddress
	if err := json.NewDecoder(resp.Body).Decode(&addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}

func (c *SupabaseClient) FetchContacts(ctx context.Context, clientIDs []string) ([]ClientContactDetail, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/client_contact_details?select=*", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var contacts []ClientContactDetail
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}

func (c *SupabaseClient) FetchBanks(ctx context.Context, clientIDs []string) ([]ClientBank, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/client_banks?select=*", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var banks []ClientBank
	if err := json.NewDecoder(resp.Body).Decode(&banks); err != nil {
		return nil, err
	}
	return banks, nil
}

func (c *SupabaseClient) FetchServicesAfter(ctx context.Context, after time.Time) ([]Service, error) {
	var endpoint string
	if after.IsZero() {
		endpoint = fmt.Sprintf("%s/rest/v1/services?select=*", c.baseURL)
	} else {
		endpoint = fmt.Sprintf("%s/rest/v1/services?select=*&updated_at=gt.%s",
			c.baseURL, after.UTC().Format("2006-01-02T15:04:05.999999Z"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var services []Service
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, err
	}
	return services, nil
}

func (c *SupabaseClient) FetchLatestMatterReference(ctx context.Context, initials string) (string, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/matters?select=reference&reference=like.WW-%s-*&deleted=eq.false&order=reference.desc&limit=10", c.baseURL, initials)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var rows []struct {
		Reference string `json:"reference"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return "", err
	}

	for _, row := range rows {
		parts := strings.Split(row.Reference, "-")
		if len(parts) == 3 {
			var parsed int
			fmt.Sscanf(parts[2], "%d", &parsed)
			if parsed < 1000000 {
				return row.Reference, nil
			}
		}
	}
	return "", nil
}

func (c *SupabaseClient) UpsertRecord(ctx context.Context, record ContactRecord) error {
	payload := map[string]interface{}{
		"id":         record.ID,
		"fields":     record.Fields,
		"updated_at": record.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999Z"),
		"deleted":    record.Deleted,
	}
	return c.UpsertGeneric(ctx, record.Category, payload)
}

func (c *SupabaseClient) UpsertGeneric(ctx context.Context, table string, payload any) error {
	endpoint := fmt.Sprintf("%s/rest/v1/%s?on_conflict=id", c.baseURL, table)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	c.setHeaders(req)
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upsert %s status %d", table, resp.StatusCode)
	}

	return nil
}

func (c *SupabaseClient) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}
