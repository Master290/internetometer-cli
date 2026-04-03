package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Master290/internetometer-cli/pkg/yandex"
)

const phaseDuration = 8.0 // seconds per download/upload phase

func main() {
	showIP := flag.Bool("ip", false, "Show IPv4 and IPv6 addresses")
	showSpeed := flag.Bool("speed", false, "Run speed test (latency, download, upload)")
	showFull := flag.Bool("all", false, "Run all tests and show full info")
	asJSON := flag.Bool("json", false, "Output results in JSON format")
	lang := flag.String("lang", "en", "Language for region (en or ru)")
	concurrency := flag.Int("concurrency", 4, "Number of concurrent connections for speed test")
	savePath := flag.String("save", "", "Path to save results in JSONL format")
	prometheus := flag.Bool("prometheus", false, "Output results in Prometheus metrics format")
	useTUI := flag.Bool("tui", false, "Use interactive TUI for progress")
	timeout := flag.Duration("timeout", 60*time.Second, "Timeout for the entire operation")
	stream := flag.Bool("stream", false, "Print one progress line per second (for non-TTY / widget use)")

	flag.Parse()

	// stream implies all + no TUI
	if *stream {
		*showFull = true
		*useTUI = false
	}

	if !*showIP && !*showSpeed && !*showFull && !*prometheus && !*asJSON && !*stream {
		*useTUI = true
	}

	client := yandex.NewClient(&yandex.Config{
		Timeout:     *timeout,
		Language:    *lang,
		Concurrency: *concurrency,
	})

	if *useTUI {
		err := yandex.RunTUI(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// -----------------------------------------------------------------------
	// Stream mode: print progress lines every second, final result at the end
	// -----------------------------------------------------------------------
	if *stream {
		// 1. Gather info up-front and print immediately
		ipv4, _ := client.GetIPv4()
		ipv6, _ := client.GetIPv6()
		region, _ := client.GetRegion()
		isp, _ := client.GetISP()

		fmt.Printf("IPv4:   %s\n", ipv4)
		if ipv6 != "" {
			fmt.Printf("IPv6:   %s\n", ipv6)
		}
		if region != "" {
			fmt.Printf("Region: %s\n", region)
		}
		if isp != nil && isp.Name != "" {
			fmt.Printf("ISP:    %s (AS%s)\n", isp.Name, isp.ASN)
		}
		fmt.Println()

		// 2. Shared state updated by progress callback
		var mu sync.Mutex
		var curPhase string
		var curMbps float64
		var curPct float64
		var phaseStart time.Time

		progressFn := func(r yandex.ProgressReport) {
			mu.Lock()
			defer mu.Unlock()

			phase := "Download"
			if !r.IsDownload {
				phase = "Upload"
			}

			// phase switch or first call
			if phase != curPhase || phaseStart.IsZero() {
				curPhase = phase
				phaseStart = time.Now()
				curMbps = 0
				curPct = 0
			}

			elapsed := time.Since(phaseStart).Seconds()
			if elapsed > 0.05 {
				curMbps = float64(r.Bytes) * 8 / (elapsed * 1e6)
				curPct = math.Min(elapsed/phaseDuration*100, 100)
			}
		}

		// 3. Ticker prints one line per second
		ticker := time.NewTicker(time.Second)
		done := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					mu.Lock()
					phase := curPhase
					mbps := curMbps
					pct := curPct
					mu.Unlock()
					if phase != "" {
						fmt.Printf("Measuring %s: %.2f Mbps (%.0f%%)\n", phase, mbps, pct)
					}
				case <-done:
					return
				}
			}
		}()

		// 4. Run speed test
		speed, err := client.RunSpeedTest(ctx, progressFn)
		ticker.Stop()
		close(done)

		fmt.Println()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Speed test failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Download: %.2f Mbps\n", speed.DownloadMbps)
		fmt.Printf("Upload:   %.2f Mbps\n", speed.UploadMbps)
		fmt.Printf("Latency:  %d ms\n", speed.Latency.Milliseconds())
		return
	}

	// -----------------------------------------------------------------------
	// Non-stream modes (original behaviour)
	// -----------------------------------------------------------------------
	results := make(map[string]interface{})

	if *showIP || *showFull {
		ipv4, err := client.GetIPv4()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting IPv4: %v\n", err)
		} else {
			results["ipv4"] = ipv4
		}

		ipv6, _ := client.GetIPv6()
		if ipv6 != "" {
			results["ipv6"] = ipv6
		}

		region, err := client.GetRegion()
		if err == nil {
			results["region"] = region
		}

		isp, _ := client.GetISP()
		if isp != nil {
			results["isp"] = isp.Name
			results["asn"] = isp.ASN
		}
	}

	if *showFull {
		results["os"] = runtime.GOOS
		results["arch"] = runtime.GOARCH
		results["num_cpu"] = runtime.NumCPU()
		results["time"] = time.Now().Format(time.RFC3339)
	}

	if *showSpeed || *showFull || *prometheus {
		if !*asJSON && !*prometheus {
			fmt.Fprintln(os.Stderr, "Running speed test...")
		}

		var startTime time.Time
		var isDownload bool = true
		progress := func(p yandex.ProgressReport) {
			if *asJSON || *prometheus {
				return
			}
			if startTime.IsZero() || p.IsDownload != isDownload {
				startTime = time.Now()
				isDownload = p.IsDownload
			}
			duration := time.Since(startTime).Seconds()
			if duration > 0 {
				mbps := (float64(p.Bytes) * 8) / (duration * 1000000.0)
				label := "Download"
				if !p.IsDownload {
					label = "Upload  "
				}
				fmt.Printf("\r%s: %.2f Mbps", label, mbps)
			}
		}

		speed, err := client.RunSpeedTest(ctx, progress)
		if err == nil && !*asJSON && !*prometheus {
			fmt.Print("\r                         \r")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Speed test failed: %v\n", err)
		} else {
			results["download_mbps"] = speed.DownloadMbps
			results["upload_mbps"] = speed.UploadMbps
			results["latency_ms"] = speed.Latency.Milliseconds()
		}
	}

	if *prometheus {
		labels := ""
		if isp, ok := results["isp"].(string); ok {
			labels += fmt.Sprintf("isp=%q,", isp)
		}
		if region, ok := results["region"].(string); ok {
			labels += fmt.Sprintf("region=%q,", region)
		}
		if len(labels) > 0 {
			labels = "{" + labels[:len(labels)-1] + "}"
		}

		fmt.Println("# HELP internetometer_download_mbps Download speed in Mbps")
		fmt.Println("# TYPE internetometer_download_mbps gauge")
		fmt.Printf("internetometer_download_mbps%s %.2f\n", labels, results["download_mbps"])

		fmt.Println("# HELP internetometer_upload_mbps Upload speed in Mbps")
		fmt.Println("# TYPE internetometer_upload_mbps gauge")
		fmt.Printf("internetometer_upload_mbps%s %.2f\n", labels, results["upload_mbps"])

		fmt.Println("# HELP internetometer_latency_ms Network latency in milliseconds")
		fmt.Println("# TYPE internetometer_latency_ms gauge")
		fmt.Printf("internetometer_latency_ms%s %v\n", labels, results["latency_ms"])
		return
	}

	if *asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(results)
	} else {
		printText(results)
	}

	if *savePath != "" {
		saveResult(results, *savePath)
	}
}

func saveResult(res map[string]interface{}, path string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return
	}
	defer f.Close()
	data, _ := json.Marshal(res)
	f.WriteString(string(data) + "\n")
}

func printText(res map[string]interface{}) {
	fmt.Println("--- Yandex Internetometer CLI ---")
	if v, ok := res["ipv4"]; ok {
		fmt.Printf("IPv4: %v\n", v)
	}
	if v, ok := res["ipv6"]; ok {
		fmt.Printf("IPv6: %v\n", v)
	} else if _, ipOk := res["ipv4"]; ipOk {
		fmt.Println("IPv6: -")
	}

	if v, ok := res["region"]; ok {
		fmt.Printf("Region:   %v\n", v)
	}
	if v, ok := res["isp"]; ok {
		fmt.Printf("ISP:      %v (AS%v)\n", v, res["asn"])
	}

	if v, ok := res["download_mbps"]; ok {
		fmt.Printf("Download: %.2f Mbps\n", v)
	}
	if v, ok := res["upload_mbps"]; ok {
		fmt.Printf("Upload:   %.2f Mbps\n", v)
	}
	if v, ok := res["latency_ms"]; ok {
		fmt.Printf("Latency:  %v ms\n", v)
	}

	if v, ok := res["os"]; ok {
		fmt.Printf("OS:       %v (%v)\n", v, res["arch"])
	}
	if v, ok := res["time"]; ok {
		fmt.Printf("Time:     %v\n", v)
	}
}
