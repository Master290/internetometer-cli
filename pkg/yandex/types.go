package yandex

type IPResponse struct {
	IP string `json:"ip"`
}

type Probe struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"`
}

type UploadProbeGroup struct {
	Duration int `json:"duration"`
	Probes   []struct {
		Paths []string `json:"urls"`
		Size  int      `json:"size"`
		Count int      `json:"count"`
	} `json:"probes"`
}

type ProbesResponse struct {
	MID     string   `json:"mid"`
	LID     []string `json:"lid"`
	Latency struct {
		Probes []Probe `json:"probes"`
	} `json:"latency"`
	Download struct {
		Probes []Probe `json:"probes"`
	} `json:"download"`
	Upload struct {
		Warmup struct {
			Duration int `json:"duration"`
			Probes   []struct {
				URLs  []string `json:"urls"`
				Size  int      `json:"size"`
				Count int      `json:"count"`
			} `json:"probes"`
		} `json:"warmup"`
		Probes []struct {
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"probes"`
	} `json:"upload"`
}

type RegionInfo struct {
	Name string `json:"name"`
}

type FullInfo struct {
	IPv4      string `json:"ipv4"`
	IPv6      string `json:"ipv6"`
	Region    string `json:"region"`
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	Timestamp string `json:"timestamp"`
}
