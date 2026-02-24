package yandex

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetIPv4() (string, error) {
	var ip string
	url := "https://ipv4-internet.yandex.net/api/v0/ip"
	err := c.get(url, &ip)
	if err != nil {
		return "", fmt.Errorf("ipv4 detection failed: %w", err)
	}
	return ip, nil
}

func (c *Client) GetIPv6() (string, error) {
	var ip string
	url := "https://ipv6-internet.yandex.net/api/v0/ip"
	err := c.get(url, &ip)
	if err != nil {
		// ipv6 might be missing, not an error
		return "", nil
	}
	return ip, nil
}

func (c *Client) GetServerTime() (string, error) {
	req, err := http.NewRequest("GET", "https://yandex.ru/internet/api/v1/datetime", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
