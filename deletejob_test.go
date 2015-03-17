package jenkins

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
        if r.Header.Get("Authorization") != "Basic dTpw" {
            t.Fatalf("Want Basic dTpw but got %s\n", r.Header.Get("Authorization"))
        }
		w.Header().Add("Location", "http://localhost:55555")
		w.WriteHeader(http.StatusFound)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	jenkinsClient := NewClient(url, "u", "p")
	err := jenkinsClient.DeleteJob("jobname")
	if err != nil {
		t.Fatalf("job-delete not expecting an error, but received: %v\n", err)
	}
}

func TestDeleteJob500(t *testing.T) {
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
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
    jenkinsClient := NewClient(url, "u", "p")
	if err := jenkinsClient.DeleteJob("jobname"); err == nil {
		t.Fatalf("job-delete expecting an error, but received none\n")
	}
}
