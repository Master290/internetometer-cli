package metrics

import (
	"sync"

	"github.com/Master290/internetometer-cli/pkg/yandex"
	"github.com/prometheus/client_golang/prometheus"
)

type internetometer struct {
	sync.RWMutex

	pingMetric     *prometheus.Desc
	uploadMetric   *prometheus.Desc
	downloadMetric *prometheus.Desc

	res yandex.SpeedResult
}

// Collect implements [prometheus.Collector].
func (i *internetometer) Collect(ch chan<- prometheus.Metric) {
	i.RLock()
	defer i.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		i.pingMetric,
		prometheus.GaugeValue,
		float64(i.res.Latency.Milliseconds()),
	)

	ch <- prometheus.MustNewConstMetric(
		i.uploadMetric,
		prometheus.GaugeValue,
		i.res.UploadMbps,
	)

	ch <- prometheus.MustNewConstMetric(
		i.downloadMetric,
		prometheus.GaugeValue,
		i.res.DownloadMbps,
	)
}

// Describe implements [prometheus.Collector].
func (i *internetometer) Describe(ch chan<- *prometheus.Desc) {
	ch <- i.pingMetric
	ch <- i.uploadMetric
	ch <- i.downloadMetric
}

func (i *internetometer) Update(res *yandex.SpeedResult) {
	i.Lock()
	defer i.Unlock()

	i.res = *res
}

func New() *internetometer {
	return &internetometer{
		pingMetric: prometheus.NewDesc(
			"internetometer_ping",
			"Latency (ms)",
			nil, nil,
		),
		uploadMetric: prometheus.NewDesc(
			"internetometer_upload",
			"Upload speed (Mb/s)",
			nil, nil,
		),

		downloadMetric: prometheus.NewDesc(
			"internetometer_download",
			"Download speed (Mb/s)",
			nil, nil,
		),
	}
}

var _ prometheus.Collector = (*internetometer)(nil)
