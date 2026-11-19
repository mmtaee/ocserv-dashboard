package system

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (u *Usecase) Release(ctx context.Context) (*DashboardRelease, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/mmtaee/ocserv-dashboard/releases/latest", nil)
	if err != nil {
		return nil, errors.New("failed to create latest release request")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "ocserv-dashboard")
	response, err := u.httpClient.Do(request)
	if err != nil {
		return nil, errors.New("failed to fetch latest release")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch latest release")
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, errors.New("failed to parse latest release")
	}
	current := u.installedRelease()
	latest := strings.TrimSpace(payload.TagName)
	return &DashboardRelease{Current: current, Latest: latest, Available: newerRelease(current, latest)}, nil
}

func (u *Usecase) installedRelease() string {
	value, err := os.ReadFile(filepath.Join(u.installDir, ".release"))
	if err == nil && strings.TrimSpace(string(value)) != "" {
		return strings.TrimSpace(string(value))
	}
	return strings.TrimSpace(u.currentRelease)
}

func newerRelease(current, latest string) bool {
	currentParts, currentOK := releaseParts(current)
	latestParts, latestOK := releaseParts(latest)
	if !currentOK || !latestOK {
		return false
	}
	for len(currentParts) < len(latestParts) {
		currentParts = append(currentParts, 0)
	}
	for len(latestParts) < len(currentParts) {
		latestParts = append(latestParts, 0)
	}
	for index := range latestParts {
		if latestParts[index] != currentParts[index] {
			return latestParts[index] > currentParts[index]
		}
	}
	return false
}

func releaseParts(value string) ([]int, bool) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	if value == "" {
		return nil, false
	}
	parts := strings.Split(value, ".")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return nil, false
		}
		result = append(result, number)
	}
	return result, true
}
