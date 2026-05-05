package yandex

import (
	"encoding/json"
	"net/http"
)

type ISPInfo struct {
	Name string `json:"name"`
	ASN  int    `json:"asn"`
}

func (c *Client) GetISP() (*ISPInfo, error) {
	req, err := http.NewRequest("GET", "http://ip-api.com/json/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		ISP string `json:"isp"`
		AS  string `json:"as"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	info := &ISPInfo{
		Name: data.ISP,
	}
	
	if len(data.AS) > 2 && data.AS[:2] == "AS" {
		var asn int
		for i := 2; i < len(data.AS); i++ {
			if data.AS[i] >= '0' && data.AS[i] <= '9' {
				asn = asn*10 + int(data.AS[i]-'0')
			} else {
				break
			}
		}
		info.ASN = asn
	}

	return info, nil
}
