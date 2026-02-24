package yandex

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type ISPInfo struct {
	Name string `json:"name"`
	ASN  int    `json:"asn"`
}

func (c *Client) GetISP() (*ISPInfo, error) {
	req, err := http.NewRequest("GET", "https://yandex.ru/internet", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	info := &ISPInfo{}

	asnRe := regexp.MustCompile(`"asn":\[(\d+)\]`)
	if match := asnRe.FindStringSubmatch(html); len(match) > 1 {
		fmt.Sscanf(match[1], "%d", &info.ASN)
	}

	ispRe := regexp.MustCompile(`"operatorName":"([^"]*)"`)
	if match := ispRe.FindStringSubmatch(html); len(match) > 1 {
		info.Name = match[1]
	}

	if info.Name == "" && info.ASN != 0 {
		info.Name = fmt.Sprintf("AS%d", info.ASN)
	}

	return info, nil
}
