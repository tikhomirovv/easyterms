package urlfetch_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tikhomirovv/easyterms/internal/ingest/urlfetch"
)

func TestNormalizeURL(t *testing.T) {
	got, err := urlfetch.NormalizeURL("https://example.com/terms")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/terms" {
		t.Fatalf("got %q", got)
	}
	_, err = urlfetch.NormalizeURL("ftp://example.com")
	if !errors.Is(err, urlfetch.ErrInvalidURL) {
		t.Fatalf("err = %v", err)
	}
}

func TestFetchText_HTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><script>ignore</script></head><body><h1>Terms</h1><p>Hello world.</p></body></html>`))
	}))
	defer srv.Close()

	text, err := urlfetch.FetchText(context.Background(), srv.URL, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if text != "Terms Hello world." {
		t.Fatalf("text = %q", text)
	}
}

func TestFetchText_unreachable(t *testing.T) {
	_, err := urlfetch.FetchText(context.Background(), "http://127.0.0.1:1/nope", &http.Client{Timeout: 0})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchText_emptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body></body></html>`))
	}))
	defer srv.Close()

	_, err := urlfetch.FetchText(context.Background(), srv.URL, srv.Client())
	if !errors.Is(err, urlfetch.ErrEmptyContent) {
		t.Fatalf("err = %v", err)
	}
}
