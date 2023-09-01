package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Configuration
const healthCheckInterval = 1 * time.Minute
const appsConfigPath = "apps.json"
const outputFilePath = "output.txt"

// App represents an application hosted on a node
type App struct {
	Domain         string `json:"domain"`
	HealthEndpoint string `json:"healthEndpoint"`
}

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

	var output string
	for host, domains := range apps {
		for _, domainConfig := range domains {
			status := "DOWN"
			resp, err := http.Get("http://" + host + domainConfig.HealthEndpoint)
			if err == nil && resp.StatusCode == 200 {
				status = "UP"
			}
			output += fmt.Sprintf("%s: %s\n", domainConfig.Domain, status)
		}
	}

	writeToFile(outputFilePath, output)
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
