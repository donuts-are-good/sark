package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

type App struct {
	Domain         string `json:"domain"`
	HealthEndpoint string `json:"healthEndpoint"`
}

type Metrics struct {
	Last200     time.Time
	LastRequest time.Time
	LastFail    time.Time
}

var domainMetrics = make(map[string]Metrics)

func main() {
	if err := loadConfigurations(); err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Duration(config.HealthCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		checkHealthAndReport()
	}
}

func loadConfigurations() error {
	file, err := os.Open("config.json")
	if err != nil {
		return fmt.Errorf("Error opening config: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return fmt.Errorf("Error decoding config: %v", err)
	}

	client.Timeout = time.Duration(config.HTTPClientTimeout) * time.Second
	return nil
}

func checkHealthAndReport() {
	apps := loadAppsConfiguration()

	var builder strings.Builder

	for host, domains := range apps {
		for _, domainConfig := range domains {
			status, currentTime := checkDomain(host, domainConfig)
			builder.WriteString(fmt.Sprintf("%s: %s | Last 200: %s | Last Request: %s | Last Fail: %s\n",
				domainConfig.Domain, status, currentTime.Last200, currentTime.LastRequest, currentTime.LastFail))
		}
	}

	writeToFile(config.OutputFilePath, builder.String())
}

func checkDomain(host string, domainConfig App) (status string, currentTime Metrics) {
	url := fmt.Sprintf("http://%s/%s", host, strings.TrimPrefix(domainConfig.HealthEndpoint, "/"))
	resp, err := client.Get(url)

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	currentTime = domainMetrics[domainConfig.Domain]
	currentTime.LastRequest = time.Now()

	if err == nil && resp != nil {
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

	domainMetrics[domainConfig.Domain] = currentTime
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
