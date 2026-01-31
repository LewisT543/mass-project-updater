package gitlab

import (
	"encoding/json"
	"net/http"
	"strings"

	"lewist543.com/mass-project-updater/internal/model"
)

func FetchUIProjects(baseURL, token, groupID, prefix string) ([]model.Project, error) {
	req, _ := http.NewRequest(
		"GET",
		baseURL+"/api/v4/groups/"+groupID+"/projects?per_page=100",
		nil,
	)
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var all []model.Project
	json.NewDecoder(resp.Body).Decode(&all)

	var filtered []model.Project
	for _, p := range all {
		if strings.HasPrefix(p.Name, prefix) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}
