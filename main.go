package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Println("Error opening config:", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Println("Error decoding config:", err)
		return
	}

	// Set client timeout
	client.Timeout = time.Duration(config.HTTPClientTimeout) * time.Second
}

type App struct {
	Domain         string `json:"domain"`
	HealthEndpoint string `json:"healthEndpoint"`
}

type Metrics struct {
	Last200     time.Time
	LastRequest time.Time
	LastFail    time.Time
}

var domainMetrics sync.Map
var wg sync.WaitGroup

func main() {
	loadConfigurations()

	ticker := time.NewTicker(time.Duration(config.HealthCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			loadConfigurations()
			checkHealthAndReport()
		}
	}
}

func checkHealthAndReport() {
	apps := loadAppsConfiguration()

	var builder strings.Builder

	for host, domains := range apps {
		wg.Add(1)
		go func(h string, doms []App) {
			defer wg.Done()
			for _, domainConfig := range doms {
				status, currentTime := checkDomain(h, domainConfig)
				// Writing the output to the file
				builder.WriteString(fmt.Sprintf("%s: %s | Last 200: %s | Last Request: %s | Last Fail: %s\n",
					domainConfig.Domain, status, currentTime.Last200, currentTime.LastRequest, currentTime.LastFail))
			}
		}(host, domains)
	}
	wg.Wait()

	writeToFile(config.OutputFilePath, builder.String())
}

func checkDomain(host string, domainConfig App) (status string, currentTime Metrics) {
	url := fmt.Sprintf("http://%s/%s", host, strings.TrimPrefix(domainConfig.HealthEndpoint, "/"))
	resp, err := client.Get(url)

	metrics, _ := domainMetrics.LoadOrStore(domainConfig.Domain, &Metrics{})
	currentTime = *metrics.(*Metrics)
	currentTime.LastRequest = time.Now()

	if err == nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		if resp.StatusCode == 200 {
			status = "UP"
			currentTime.Last200 = time.Now()
		} else {
			status = "DOWN"
			currentTime.LastFail = time.Now()
		}
	} else {
		log.Println("Error checking domain:", domainConfig.Domain, err)
		status = "DOWN"
		currentTime.LastFail = time.Now()
	}

	domainMetrics.Store(domainConfig.Domain, &currentTime)
	return
}

func loadAppsConfiguration() map[string][]App {
	file, err := os.Open(config.AppsConfigPath)
	if err != nil {
		log.Println("Error opening apps config:", err)
		return nil
	}
	defer file.Close()

	var apps map[string][]App
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&apps)
	if err != nil {
		log.Println("Error decoding apps config:", err)
		return nil
	}
	return apps
}

func writeToFile(filePath, data string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		log.Println("Error writing to file:", err)
	}
}
