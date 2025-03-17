package monitor

import (
	"ChintuIdrive/s3-watchdog/clients"
	"ChintuIdrive/s3-watchdog/collector"
	"ChintuIdrive/s3-watchdog/conf"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func StartMonitor(config *conf.Config) {
	asc := clients.NewApiServerClient(config)
	err := asc.Login()
	if err != nil {
		log.Printf("Failed to log in to api server: %v", err)
		return
	}
	s3mc := collector.NewS3MetricCollector(config)
	s3monitor := NewS3StatsMonitor(config, s3mc, asc)
	regions, err := asc.GetRegions()
	if err != nil {
		//notify watchdog not able to fetch tenantlist from api server
		log.Printf("Failed to fetch regions from api server: %v", err)
		return
	}
	//var wg sync.WaitGroup

	for _, region := range regions {
		if region.Region == config.Region {
			nodes := strings.Split(region.StorageNodes, ",")
			for _, node := range nodes {
				log.Printf("Starting S3 monitoring for node: %s", node)
				go s3monitor.MonitorTenantsS3Stats(node)
				// if(node=="or1"){
					
				// }
				// wg.Add(1) // Increase wait group counter
				// go func(n string) {
				// 	defer wg.Done()
				// 	s3monitor.MonitorTenantsS3Stats(n)
				// }(node)
			}
		}
	}
	router := gin.Default()
	//router.POST("/api/certificate/create", co.CreateCertificate)
	router.Run(":8786")
	//wg.Wait() // Ensures main function waits for goroutines to finish
}
