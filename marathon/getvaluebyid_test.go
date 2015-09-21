package marathon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetValueById(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note the path is currently hard-wired and does not reflect the requested key
		if r.URL.Path != "/v1/kv/priority1-demand" {
			t.Fatalf("Url not as expected, have %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Fatalf("Expected GET, have %s", r.Method)
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[
   {
       "CreateIndex": 8,
       "ModifyIndex": 15,
       "LockIndex": 0,
       "Key": "priority1-demand",
       "Flags": 0,
       "Value": "myvalue"
   }
]`)
	}))
	defer server.Close()

	m := NewMarathonScheduler()
	m.baseConsulUrl = server.URL

	value, err := m.GetValuebyID("hat")
	if err != nil {
		t.Fatalf("GetValuebyID returned an error %v", err)
	}

	if value != "myvalue" {
		t.Fatalf("Value not as expected. Have %s", value)
	}
}
