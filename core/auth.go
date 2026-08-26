package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) AuthenticateUser(username, password string) error {
	endpoint := fmt.Sprintf("%s/rest/v1/users?username=eq.%s&select=*", a.supabase.baseURL, url.QueryEscape(username))
	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create auth request: %w", err)
	}
	a.supabase.setHeaders(req)

	resp, err := a.supabase.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication service unavailable")
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return fmt.Errorf("invalid authentication response")
	}

	if len(users) == 0 {
		return fmt.Errorf("invalid credentials")
	}

	user := users[0]
	if !user.Enabled {
		return fmt.Errorf("account disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return fmt.Errorf("invalid credentials")
	}

	_ = a.SaveLastUsername(username)
	return nil
}

func (a *App) SaveLastUsername(username string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "LuminaLex", "config.json")

	config := AppConfig{LastUsername: username}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func (a *App) GetLastUsername() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	configPath := filepath.Join(configDir, "LuminaLex", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}
	return config.LastUsername
}
