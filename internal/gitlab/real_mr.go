package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type RealMRClient struct {
	BaseURL string
	Token   string
}

func NewMRClient(baseURL, token string) MRClient {
	return &RealMRClient{BaseURL: baseURL, Token: token}
}

func (c *RealMRClient) CreateMR(projectID int, sourceBranch, targetBranch, title string) (string, error) {
	payload := map[string]string{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"title":         title,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests", c.BaseURL, projectID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("PRIVATE-TOKEN", c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		WebURL string `json:"web_url"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.WebURL, nil
}
