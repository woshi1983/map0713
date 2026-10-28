package avi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAVISelectsControllerSupportedAPIVersion(t *testing.T) {
	for _, tc := range []struct {
		name       string
		controller string
		minimum    string
		want       string
	}{
		{name: "old controller", controller: "20.1.6", minimum: "20.1.3", want: "20.1.6"},
		{name: "new controller", controller: "31.1.2", minimum: "20.1.7", want: "20.1.7"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			versionCalls := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/cluster/version":
					versionCalls++
					user, password, ok := r.BasicAuth()
					if !ok || user != "test-user" || password != "test-password" {
						t.Errorf("version discovery did not authenticate")
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					fmt.Fprintf(w, `{"Version":%q,"min_version":%q}`, tc.controller, tc.minimum)
				case "/login", "/api/serviceenginegroup", "/logout":
					if got := r.Header.Get("X-Avi-Version"); got != tc.want {
						t.Errorf("%s used version %q; want %q", r.URL.Path, got, tc.want)
					}
					if r.URL.Path == "/api/serviceenginegroup" {
						fmt.Fprint(w, `{"count":0,"results":[]}`)
					}
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client, err := NewAVI(server.URL, "test-user", []byte("test-password"), true)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				if _, err := client.ListSEGroupContext(context.Background(), &ListSEGroupOptions{Name: "Default-Group"}); err != nil {
					t.Fatal(err)
				}
			}
			if versionCalls != 1 {
				t.Fatalf("version discovery called %d times; want once", versionCalls)
			}
		})
	}
}
