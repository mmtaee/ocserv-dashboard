package system

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	return &DashboardRelease{Current: u.currentRelease, Latest: strings.TrimSpace(payload.TagName)}, nil
}
