package nichan_test

import (
	"."
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrawl(t *testing.T) {
	html := `
<html>
<body>
<h1>Test Thread Title</h1>
</body>
</htm>
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, html)
	}))
	defer ts.Close()

	th, err := nichan.Crawl(ts.URL)
	if err != nil {
		t.Errorf("failed to Crawl: %v", err)
	}

	if th.Title != "Test Thread Title" {
		t.Errorf("wrong title")
	}
}

func TestCrawlHttpError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found.", http.StatusNotFound)
	}))
	defer ts.Close()

	th, err := nichan.Crawl(ts.URL)
	if err == nil {
		t.Errorf("we expect to fail http.Get")
	}
	fmt.Println(th)
}
