package core

import "time"

type ContactRecord struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Fields    []string  `json:"fields"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Synced    bool      `json:"synced"`
}

type UpdateCheckResult struct {
	HasUpdate    bool   `json:"has_update"`
	LatestVer    string `json:"latest_ver"`
	ReleaseNotes string `json:"release_notes"`
	DownloadURL  string `json:"download_url"`
}

type SyncStatus struct {
	IsSyncing bool   `json:"is_syncing"`
	LastSync  string `json:"last_sync"`
	Error     string `json:"error"`
}
