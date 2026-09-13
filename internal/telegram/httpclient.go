package telegram

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// TelegramPollTimeout matches go-telegram/bot default long-polling timeout.
const TelegramPollTimeout = time.Minute

// NewBotHTTPClient builds an HTTP client for the Telegram Bot API.
// proxyURL is optional; supports http(s):// and socks5/socks5h:// (remote DNS).
func NewBotHTTPClient(proxyURL string) (*http.Client, error) {
	transport := baseTransport()

	raw := strings.TrimSpace(proxyURL)
	if raw == "" {
		return &http.Client{
			Timeout:   TelegramPollTimeout + 15*time.Second,
			Transport: transport,
		}, nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("TELEGRAM_PROXY: invalid URL: %w", err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("TELEGRAM_PROXY: missing host")
	}

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "http", "https":
		transport.Proxy = http.ProxyURL(u)
	case "socks5", "socks5h":
		// socks5h: hostname resolved by the proxy (same as curl socks5h).
		socksURL := *u
		socksURL.Scheme = "socks5"
		dialer, err := proxy.FromURL(&socksURL, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("TELEGRAM_PROXY: %w", err)
		}
		transport.Proxy = nil
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if ctxDialer, ok := dialer.(proxy.ContextDialer); ok {
				return ctxDialer.DialContext(ctx, network, addr)
			}
			return dialer.Dial(network, addr)
		}
	default:
		return nil, fmt.Errorf("TELEGRAM_PROXY: unsupported scheme %q (use http, https, socks5, socks5h)", u.Scheme)
	}

	return &http.Client{
		Timeout:   TelegramPollTimeout + 15*time.Second,
		Transport: transport,
	}, nil
}

// RedactProxyURL returns a log-safe proxy URL (password masked).
func RedactProxyURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid>"
	}
	if u.User != nil {
		user := u.User.Username()
		if _, hasPass := u.User.Password(); hasPass {
			u.User = url.UserPassword(user, "***")
		}
	}
	return u.String()
}

func baseTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
