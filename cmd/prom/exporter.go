package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Master290/internetometer-cli/cmd/prom/metrics"
	"github.com/Master290/internetometer-cli/pkg/yandex"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	delayStr := flag.String("delay", "1h", "Delay between measurements (in time.Duration format)")
	timeoutStr := flag.String("timeout", "60s", "Timeout for measurement operation")
	iface := flag.String("interface", "", "Network interface to bind to")
	flag.Parse()

	if d, exists := os.LookupEnv("IM_DELAY"); exists {
		*delayStr = d
	}

	if to, exists := os.LookupEnv("IM_TIMEOUT"); exists {
		*timeoutStr = to
	}

	if i, exists := os.LookupEnv("IM_INTERFACE"); exists {
		*iface = i
	}

	delay, err := time.ParseDuration(*delayStr)
	if err != nil {
		log.Fatal(err)
	}

	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		log.Fatal(err)
	}

	client := yandex.NewClient(&yandex.Config{
		Timeout:     timeout,
		Concurrency: 1,
		Interface:   *iface,
	})

	m := metrics.New()
	prometheus.MustRegister(m)

	go func() {
		ticker := time.NewTicker(delay)
		for {
			log.Println("Measuring Internet connectivity parameters.")

			speed, err := client.RunSpeedTest(context.TODO(), nil)
			if err != nil {
				log.Println(err)
			} else {
				m.Update(speed)
			}

			log.Println("Background cache updated.")
			<-ticker.C
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9112", nil))

}
