package jenkins

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetLastBuild(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := *r.URL
		if url.Path != "/job/thejob/lastBuild/api/json" {
			t.Fatalf("Want /job/thejob/lastBuild/api/json but got %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("Want application/json but got %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want Basic dTpw but got %s\n", r.Header.Get("Authorization"))
		}
		w.Write([]byte(`{"result":"SUCCESS","timestamp":1456425493292,"url":"https://server/job/thejob/1/"}`))
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	jenkinsClient := NewClient(url, "u", "p")
	lastBuild, err := jenkinsClient.GetLastBuild("thejob")
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	if lastBuild.Result != "SUCCESS" {
		t.Fatalf("Want SUCCESS but got: %s\n", lastBuild.Result)
	}
	if lastBuild.TimestampMillis != 1456425493292 {
		t.Fatalf("Want 1456425493292 but got: %d\n", lastBuild.TimestampMillis)
	}
	if lastBuild.URL != "https://server/job/thejob/1/" {
		t.Fatalf("Want https://server/job/thejob/1/ but got: %s\n", lastBuild.URL)
	}
}
