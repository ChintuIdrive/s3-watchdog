package main

import (
	"ChintuIdrive/s3-watchdog/api"
	"ChintuIdrive/s3-watchdog/clients"
	"ChintuIdrive/s3-watchdog/collector"
	"ChintuIdrive/s3-watchdog/conf"
	"ChintuIdrive/s3-watchdog/monitor"
	"encoding/json"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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
	asc := clients.NewApiServerClient(config)
	//err := asc.Login()
	// if err != nil {
	// 	log.Printf("Failed to log in to api server: %v", err)
	// 	return
	// }
	s3mc := collector.NewS3MetricCollector(config)
	s3monitor := monitor.NewS3StatsMonitor(config, s3mc, asc)
	monitor.StartMonitor(config, asc, s3monitor)

	handler := api.NewHandler(s3monitor)

	router := gin.Default()
	router.POST("/api/s3-monitor/getStats", handler.GetS3Metric)
	router.Run(":8786")

}
