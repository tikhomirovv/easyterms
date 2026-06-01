// Package urlfetch validates URLs, fetches pages, and extracts plain text from HTML.
package urlfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	maxBodyBytes = 2 << 20 // 2 MiB
	userAgent    = "EasyTermsBot/1.0 (+https://github.com/tikhomirovv/easyterms)"
)

var (
	// ErrInvalidURL is returned when the URL is missing scheme/host or uses a disallowed scheme.
	ErrInvalidURL = errors.New("invalid url")
	// ErrUnreachable is returned when the HTTP request fails.
	ErrUnreachable = errors.New("url unreachable")
	// ErrEmptyContent is returned when no meaningful text could be extracted.
	ErrEmptyContent = errors.New("page has no extractable text")
)

// NormalizeURL parses and validates an http(s) URL.
func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidURL
	}
	if u.Host == "" {
		return "", ErrInvalidURL
	}
	return u.String(), nil
}

// FetchText downloads a page and returns extracted plain text.
func FetchText(ctx context.Context, rawURL string, client *http.Client) (string, error) {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return "", err
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: status %d", ErrUnreachable, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return "", fmt.Errorf("%w: read body: %v", ErrUnreachable, err)
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	text := string(body)
	if strings.Contains(ct, "html") || strings.Contains(ct, "xhtml") || looksLikeHTML(body) {
		text, err = HTMLToText(string(body))
		if err != nil {
			return "", err
		}
	}

	text = collapseWhitespace(strings.TrimSpace(text))
	if text == "" {
		return "", ErrEmptyContent
	}
	return text, nil
}

func looksLikeHTML(b []byte) bool {
	s := strings.ToLower(string(b[:min(len(b), 512)]))
	return strings.Contains(s, "<html") || strings.Contains(s, "<body")
}

// HTMLToText extracts visible text from an HTML document.
func HTMLToText(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}
	var b strings.Builder
	walkText(doc, &b)
	return collapseWhitespace(strings.TrimSpace(b.String())), nil
}

func walkText(n *html.Node, out *strings.Builder) {
	if n.Type == html.TextNode {
		t := strings.TrimSpace(n.Data)
		if t != "" {
			if out.Len() > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(t)
		}
		return
	}
	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "script", "style", "noscript", "head":
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkText(c, out)
	}
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
