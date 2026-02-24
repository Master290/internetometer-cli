package yandex

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
)

func (c *Client) GetRegion() (string, error) {
	baseURL := "https://yandex.ru/internet"
	if c.config.Language == "en" {
		baseURL = "https://yandex.com/internet"
	}

	req, err := http.NewRequest("GET", baseURL, nil)
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

	re := regexp.MustCompile(`"clientRegion":({[^}]*})`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return "Unknown", nil
	}

	var info RegionInfo
	err = json.Unmarshal(matches[1], &info)
	if err != nil {
		return "Unknown", nil
	}

	return info.Name, nil
}
