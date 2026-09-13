package telegram

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNewBotHTTPClient_noProxy(t *testing.T) {
	c, err := NewBotHTTPClient("")
	if err != nil {
		t.Fatal(err)
	}
	if c.Timeout <= TelegramPollTimeout {
		t.Fatalf("timeout too short: %v", c.Timeout)
	}
}

func TestNewBotHTTPClient_httpProxy(t *testing.T) {
	c, err := NewBotHTTPClient("http://127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Fatal("expected HTTP proxy on transport")
	}
	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "api.telegram.org"}}
	proxyURL, err := tr.Proxy(req)
	if err != nil {
		t.Fatal(err)
	}
	if proxyURL.Host != "127.0.0.1:8080" {
		t.Fatalf("proxy host %q", proxyURL.Host)
	}
}

func TestNewBotHTTPClient_socks5h(t *testing.T) {
	c, err := NewBotHTTPClient("socks5h://user:secret@proxy.example:1080")
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok || tr.DialContext == nil {
		t.Fatal("expected SOCKS dialer on transport")
	}
	if tr.Proxy != nil {
		t.Fatal("SOCKS transport should not set HTTP Proxy")
	}
}

func TestNewBotHTTPClient_invalid(t *testing.T) {
	if _, err := NewBotHTTPClient("ftp://bad"); err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestRedactProxyURL(t *testing.T) {
	got := RedactProxyURL("socks5h://easyterms-orangepi:sekret@88.210.13.78:3839")
	if got == "" || strings.Contains(got, "sekret") {
		t.Fatalf("redacted = %q", got)
	}
	if !strings.Contains(got, "easyterms-orangepi") || !strings.Contains(got, "88.210.13.78:3839") {
		t.Fatalf("redacted = %q", got)
	}
}
