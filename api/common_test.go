package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/force12io/force12/demand"
)

func TestGetBaseUrl(t *testing.T) {
	base := getBaseF12APIUrl()
	if base != "http://app.force12.io" || base != baseF12APIUrl {
		t.Fatalf("Maybe F12_METRICS_API_ADDRESS is set: %v | %v", base, baseF12APIUrl)
	}
}

// Utility for checking GET requests
func doTestGetJson(t *testing.T, expUrl string, success bool, testJson string) (server *httptest.Server) {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expUrl {
			t.Fatalf("Expected %s, have %s", expUrl, r.URL.Path)
		}

		if r.Method != "GET" {
			t.Fatalf("expected GET, have %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content type not as expected, have %s", r.Header.Get("Content-Type"))
		}

		if success {
			w.Write([]byte(testJson))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	return server
}

// Utility for checking that tasks are updated to be what we expect
func checkReturnedTasks(t *testing.T, tasks map[string]demand.Task, returned_tasks map[string]demand.Task) {
	for name, rt := range returned_tasks {
		tt, ok := tasks[name]
		if !ok {
			t.Fatalf("Unexpected app name %v", name)
		}

		if tt.Image != rt.Image {
			t.Fatalf("Image: expected %s got %s", tt.Image, rt.Image)
		}
		if tt.Command != rt.Command {
			t.Fatalf("Command: expected %s got %s", tt.Command, rt.Command)
		}
		if tt.Demand != rt.Demand {
			t.Fatalf("Demand: expected %s got %s", tt.Demand, rt.Demand)
		}
		if tt.Requested != rt.Requested {
			t.Fatalf("Requested: expected %s got %s", tt.Requested, rt.Requested)
		}
		if tt.Running != rt.Running {
			t.Fatalf("Requested: expected %s got %s", tt.Requested, rt.Requested)
		}
	}
}
