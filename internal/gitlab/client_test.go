package gitlab

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchUIProjects_FiltersCorrectly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/groups/42/projects", r.URL.Path)
		assert.Equal(t, "token123", r.Header.Get("PRIVATE-TOKEN"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
          {
            "id": 1,
            "name": "ui-spa-dashboard",
            "ssh_url_to_repo": "git@gitlab.com:acme/ui-spa-dashboard.git"
          },
          {
            "id": 2,
            "name": "api-users",
            "ssh_url_to_repo": "git@gitlab.com:acme/api-users.git"
          }
        ]`))
	}))
	defer server.Close()

	projects, err := FetchUIProjects(
		server.URL,
		"token123",
		"42",
		"ui-spa",
	)

	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "ui-spa-dashboard", projects[0].Name)
}
