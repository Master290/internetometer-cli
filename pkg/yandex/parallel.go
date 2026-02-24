package yandex

import (
	"context"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func (c *Client) measureDownloadParallel(ctx context.Context, url string, concurrency int, progress ProgressFunc) (float64, error) {
	const targetDuration = 8 * time.Second
	c.lastTestStart = time.Now()
	start := c.lastTestStart
	var totalRead int64

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if time.Since(start) >= targetDuration {
					return
				}

				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					return
				}
				req.Header.Set("User-Agent", c.config.UserAgent)
				req.Header.Set("Referer", "https://yandex.ru/")

				resp, err := c.httpClient.Do(req)
				if err != nil {
					return
				}
				if resp == nil || resp.Body == nil {
					return
				}
				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					if resp.StatusCode == http.StatusForbidden {
						time.Sleep(1 * time.Second)
					}
					continue
				}

				pr := &progressReader{
					Reader: resp.Body,
					OnRead: func(n int) {
						newTotal := atomic.AddInt64(&totalRead, int64(n))
						if progress != nil {
							progress(ProgressReport{Bytes: newTotal, IsDownload: true})
						}
					},
				}

				n, _ := io.Copy(io.Discard, pr)
				resp.Body.Close()
				if n == 0 {
					time.Sleep(200 * time.Millisecond)
				}
			}
		}()
	}

	timer := time.NewTimer(targetDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()

	duration := time.Since(start).Seconds()
	bitsPerSec := (float64(totalRead) * 8) / duration
	return bitsPerSec, nil
}

func (c *Client) measureUploadParallel(ctx context.Context, url string, size int, concurrency int, progress ProgressFunc) (float64, error) {
	const targetDuration = 8 * time.Second
	c.lastTestStart = time.Now()
	start := c.lastTestStart
	var totalWritten int64

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if time.Since(start) >= targetDuration {
					return
				}

				pw := &progressReader{
					Reader: &io.LimitedReader{R: &nullReader{}, N: int64(size)},
					OnRead: func(n int) {
						newTotal := atomic.AddInt64(&totalWritten, int64(n))
						if progress != nil {
							progress(ProgressReport{Bytes: newTotal, IsDownload: false})
						}
					},
				}

				req, err := http.NewRequestWithContext(ctx, "POST", url, pw)
				if err != nil {
					return
				}
				req.Header.Set("User-Agent", c.config.UserAgent)
				req.Header.Set("Referer", "https://yandex.ru/internet")
				req.Header.Set("Origin", "https://yandex.ru")
				req.Header.Set("Accept", "*/*")
				req.Header.Set("Sec-Fetch-Mode", "cors")
				req.Header.Set("Sec-Fetch-Site", "cross-site")
				req.Header.Set("Sec-Fetch-Dest", "empty")
				req.ContentLength = int64(size)

				resp, err := c.httpClient.Do(req)
				if err != nil {
					return
				}
				if resp == nil {
					return
				}
				if resp.Body == nil {
					return
				}
				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					time.Sleep(500 * time.Millisecond)
					continue
				}
				resp.Body.Close()
			}
		}()
	}

	timer := time.NewTimer(targetDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()

	duration := time.Since(start).Seconds()
	bitsPerSec := (float64(totalWritten) * 8) / duration
	return bitsPerSec, nil
}
