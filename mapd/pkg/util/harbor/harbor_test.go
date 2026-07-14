package harbor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetChartReposContextReturnsEmptyListOnMissingChartRepo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chartrepo/ai/charts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := NewV2Client(server.URL, "", nil, false)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := client.GetChartReposContext(context.Background(), "ai")
	if err != nil {
		t.Fatalf("GetChartReposContext returned error: %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected empty chart repo list, got %d", len(repos))
	}
}
