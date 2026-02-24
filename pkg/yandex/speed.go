package yandex

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

func (c *Client) GetProbes() (*ProbesResponse, error) {
	var resp ProbesResponse
	url := "https://yandex.ru/internet/api/v0/get-probes"
	err := c.get(url, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get probes: %w", err)
	}
	return &resp, nil
}

type SpeedResult struct {
	DownloadMbps float64
	UploadMbps   float64
	Latency      time.Duration
}

type ProgressReport struct {
	Bytes      int64
	IsDownload bool
}

type ProgressFunc func(ProgressReport)

func (c *Client) RunSpeedTest(ctx context.Context, progress ProgressFunc) (*SpeedResult, error) {
	probes, err := c.GetProbes()
	if err != nil {
		return nil, err
	}

	result := &SpeedResult{}

	// latency
	if len(probes.Latency.Probes) > 0 {
		latency, err := c.measureLatency(ctx, probes.Latency.Probes)
		if err == nil {
			result.Latency = latency
		}
	}

	// download
	if len(probes.Download.Probes) > 0 {
		var targetProbe *Probe
		for i := range probes.Download.Probes {
			p := &probes.Download.Probes[i]
			if targetProbe == nil || (p.URL != "" && !regexp.MustCompile(`100kb`).MatchString(p.URL)) {
				targetProbe = p
				if regexp.MustCompile(`50mb`).MatchString(p.URL) {
					break
				}
			}
		}
		if targetProbe == nil {
			targetProbe = &probes.Download.Probes[0]
		}

		c.lastTestStart = time.Now()
		bitsPerSec, err := c.measureDownloadParallel(ctx, targetProbe.URL, c.config.Concurrency, progress)
		if err == nil {
			result.DownloadMbps = bitsPerSec / 1000000.0
		}
	}

	// upload
	if len(probes.Upload.Probes) > 0 {
		targetURL := probes.Upload.Probes[0].URL
		if targetURL != "" {
			c.lastTestStart = time.Now()
			bitsPerSec, err := c.measureUploadParallel(ctx, targetURL, 50*1024*1024, c.config.Concurrency, progress)
			if err == nil {
				result.UploadMbps = bitsPerSec / 1000000.0
			}
		}
	}

	return result, nil
}

func (c *Client) measureLatency(ctx context.Context, probes []Probe) (time.Duration, error) {
	const count = 3
	var min time.Duration = 10 * time.Second
	found := false

	for i := 0; i < count; i++ {
		for _, p := range probes {
			if p.URL == "" {
				continue
			}
			start := time.Now()
			req, err := http.NewRequestWithContext(ctx, "GET", p.URL, nil)
			if err != nil {
				continue
			}
			req.Header.Set("User-Agent", c.config.UserAgent)
			req.Header.Set("Referer", "https://yandex.ru/internet")

			resp, err := c.httpClient.Do(req)
			if err != nil {
				continue
			}
			if resp.StatusCode == http.StatusOK {
				dur := time.Since(start)
				if dur < min {
					min = dur
				}
				found = true
			}
			resp.Body.Close()
		}
	}
	if !found {
		return 0, fmt.Errorf("no successful latency probes")
	}
	return min, nil
}

func (c *Client) measureDownload(ctx context.Context, url string, progress ProgressFunc) (float64, error) {
	const targetDuration = 8 * time.Second
	start := time.Now()
	var totalRead int64

	for time.Since(start) < targetDuration {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if totalRead > 0 {
				break
			}
			return 0, err
		}

		pr := &progressReader{
			Reader: resp.Body,
			OnRead: func(n int) {
				totalRead += int64(n)
				if progress != nil {
					progress(ProgressReport{Bytes: totalRead, IsDownload: true})
				}
			},
		}

		_, err = io.Copy(io.Discard, pr)
		resp.Body.Close()
		if err != nil {
			break
		}

		if time.Since(start) >= targetDuration {
			break
		}
	}

	duration := time.Since(start).Seconds()
	bitsPerSec := (float64(totalRead) * 8) / duration
	return bitsPerSec, nil
}

func (c *Client) measureUpload(ctx context.Context, url string, size int, progress ProgressFunc) (float64, error) {
	const targetDuration = 8 * time.Second
	start := time.Now()
	var totalWritten int64

	for time.Since(start) < targetDuration {
		pw := &progressReader{
			Reader: &io.LimitedReader{R: &nullReader{}, N: int64(size)},
			OnRead: func(n int) {
				totalWritten += int64(n)
				if progress != nil {
					progress(ProgressReport{Bytes: totalWritten, IsDownload: false})
				}
			},
		}

		req, _ := http.NewRequestWithContext(ctx, "POST", url, pw)
		req.ContentLength = int64(size)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if totalWritten > 0 {
				break
			}
			return 0, err
		}
		resp.Body.Close()

		if time.Since(start) >= targetDuration {
			break
		}
	}

	duration := time.Since(start).Seconds()
	bitsPerSec := (float64(totalWritten) * 8) / duration
	return bitsPerSec, nil
}

type progressReader struct {
	io.Reader
	OnRead func(int)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 && r.OnRead != nil {
		r.OnRead(n)
	}
	return n, err
}

type nullReader struct{}

func (r *nullReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
