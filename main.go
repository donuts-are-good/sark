package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Configuration
const healthCheckInterval = 1 * time.Minute
const appsConfigPath = "apps.json"
const outputFilePath = "output.txt"

var client = &http.Client{
	Timeout: time.Second * 5,
}

// App represents an application hosted on a node
type App struct {
	Domain         string `json:"domain"`
	HealthEndpoint string `json:"healthEndpoint"`
}

// Metrics tracks the different timestamps for an application domain
type Metrics struct {
	Last200     time.Time
	LastRequest time.Time
	LastFail    time.Time
}

var domainMetrics sync.Map

func main() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkHealthAndReport()
		}
	}
}

func checkHealthAndReport() {
	apps := loadAppsConfiguration()

	// Using strings.Builder for efficient string concatenation
	var builder strings.Builder

	for host, domains := range apps {
		for _, domainConfig := range domains {
			status := "DOWN"
			currentTime := time.Now()
			resp, err := client.Get("http://" + host + domainConfig.HealthEndpoint)

			// Initialize metrics if not existent and also load them in one step
			metrics, _ := domainMetrics.LoadOrStore(domainConfig.Domain, &Metrics{})
			metricsPtr := metrics.(*Metrics)
			metricsPtr.LastRequest = currentTime

			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					status = "UP"
					metricsPtr.Last200 = currentTime
				} else {
					metricsPtr.LastFail = currentTime
				}
			} else {
				metricsPtr.LastFail = currentTime
			}

			// Writing the output to the file
			builder.WriteString(fmt.Sprintf("%s: %s | Last 200: %s | Last Request: %s | Last Fail: %s\n",
				domainConfig.Domain, status, metricsPtr.Last200, metricsPtr.LastRequest, metricsPtr.LastFail))
		}
	}

	writeToFile(outputFilePath, builder.String())
}

func loadAppsConfiguration() map[string][]App {
	file, err := os.Open(appsConfigPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var apps map[string][]App
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&apps)
	if err != nil {
		panic(err)
	}
	return apps
}

func writeToFile(filePath, data string) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		panic(err)
	}
}
