package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type AutoUpdater struct {
	owner          string
	repo           string
	currentVersion string
	httpClient     *http.Client
}

func NewAutoUpdater(owner, repo, currentVer string) *AutoUpdater {
	return &AutoUpdater{
		owner:          owner,
		repo:           repo,
		currentVersion: currentVer,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (u *AutoUpdater) CheckForUpdates(ctx context.Context) (*UpdateCheckResult, error) {
	// 1. Safety Check: If environment variables are missing, silently disable updates.
	if u.owner == "" || u.repo == "" {
		return &UpdateCheckResult{HasUpdate: false}, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create req: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	// 2. Safety Check: If no releases exist yet, GitHub returns 404. Handle gracefully.
	if resp.StatusCode == http.StatusNotFound {
		return &UpdateCheckResult{HasUpdate: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	if release.TagName == u.currentVersion {
		return &UpdateCheckResult{HasUpdate: false}, nil
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if runtime.GOOS == "windows" && filepath.Ext(asset.Name) == ".exe" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	return &UpdateCheckResult{
		HasUpdate:    true,
		LatestVer:    release.TagName,
		ReleaseNotes: release.Body,
		DownloadURL:  downloadURL,
	}, nil
}

func (u *AutoUpdater) ApplyUpdate(ctx context.Context, downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("no executable asset download URL provided")
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("create download req: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download asset failed status %d", resp.StatusCode)
	}

	oldExePath := exePath + ".old"
	_ = os.Remove(oldExePath)

	if err := os.Rename(exePath, oldExePath); err != nil {
		return fmt.Errorf("rename running binary: %w", err)
	}

	newFile, err := os.OpenFile(exePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("create new binary file: %w", err)
	}

	_, err = io.Copy(newFile, resp.Body)
	_ = newFile.Close()
	if err != nil {
		_ = os.Remove(exePath)
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("write new binary: %w", err)
	}

	return nil
}

func (u *AutoUpdater) RestartApp() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("restart binary: %w", err)
	}
	os.Exit(0)
	return nil
}
