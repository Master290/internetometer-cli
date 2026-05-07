package yandex

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Config struct {
	BaseURL     string
	UserAgent   string
	Timeout     time.Duration
	Language    string // "ru" or "en"
	Concurrency int
	Interface   string
}

type Client struct {
	httpClient    *http.Client
	config        *Config
	lastTestStart time.Time
}

func NewClient(cfg *Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://yandex.ru/internet"
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.IdleConnTimeout = 90 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.ExpectContinueTimeout = 1 * time.Second

	if cfg.Interface != "" {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialContext(ctx, network, addr, cfg.Interface)
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		config: cfg,
	}
}

func (c *Client) get(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status: %s, body: %s", resp.Status, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) getWithContext(ctx context.Context, url string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status: %s, body: %s", resp.Status, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
