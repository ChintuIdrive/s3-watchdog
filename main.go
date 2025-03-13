package main

import (
	"ChintuIdrive/s3-watchdog/conf"
	"ChintuIdrive/s3-watchdog/monitor"
	"encoding/json"
	"log"
	"os"
)

// Log file setup
var (
	logFile *os.File
	config  *conf.Config
)

func init() {
	var err error
	if _, err := os.Stat("conf/config.json"); os.IsNotExist(err) {
		config = conf.GetDefaultConfig()
		configFile, err := os.Create("conf/config.json")
		if err != nil {
			log.Fatalf("Failed to create config file: %s", err)
		}
		defer configFile.Close()

		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal default config: %s", err)
		}
		configFile.Write(configData)

		// encoder := json.NewEncoder(configFile)
		// if err := encoder.Encode(defaultConfig); err != nil {
		// 	log.Fatalf("Failed to write default config to file: %s", err)
		// }
	} else {
		config, err = conf.LoadConfig("conf/config.json")
		if err != nil {
			log.Fatalf("Failed to load config: %s", err)
		}
	}

	logFile, err = os.OpenFile("watchdog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
func main() {
	log.Println("Hello World")
	monitor.StartMonitor(config)
}
