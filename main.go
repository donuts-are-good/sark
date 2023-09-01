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

type Configuration struct {
	HealthCheckInterval int    `json:"healthCheckInterval"`
	AppsConfigPath      string `json:"appsConfigPath"`
	OutputFilePath      string `json:"outputFilePath"`
	HTTPClientTimeout   int    `json:"httpClientTimeout"`
}

var config Configuration
var client = &http.Client{}

func loadConfigurations() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	// Set client timeout
	client.Timeout = time.Duration(config.HTTPClientTimeout) * time.Second
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
	loadConfigurations()

	ticker := time.NewTicker(time.Duration(config.HealthCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Reload configurations
			loadConfigurations()
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

	writeToFile(config.OutputFilePath, builder.String())
}

func loadAppsConfiguration() map[string][]App {
	file, err := os.Open(config.AppsConfigPath)
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
