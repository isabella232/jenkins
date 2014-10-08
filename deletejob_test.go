package jenkins

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteJob(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := r.URL
		if url.Path != "/job/jobname/doDelete" {
			t.Fatalf("wanted URL path /job/thejob/doDelete but found %s\n", url.Path)
		}
		if r.Header.Get("Content-type") != "application/xml" {
			t.Fatalf("wanted  Content-type header application/xml but found %s\n", r.Header.Get("Content-type"))
		}
		w.Header().Add("Location", "http://localhost:55555")
		w.WriteHeader(http.StatusFound)
	}))
	defer testServer.Close()

	err := DeleteJob(testServer.URL, "jobname")
	if err != nil {
		t.Fatalf("job-delete not expecting an error, but received: %v\n", err)
	}
}
