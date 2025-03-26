package monitor

import (
	"ChintuIdrive/s3-watchdog/clients"
	"ChintuIdrive/s3-watchdog/conf"
	"log"
	"strings"
)

func StartMonitor(config *conf.Config, asc *clients.APIserverClient, s3monitor *S3StatsMonitor) {

	regions, err := asc.GetRegions()
	if err != nil {
		//notify watchdog not able to fetch tenantlist from api server
		log.Printf("Failed to fetch regions from api server: %v", err)
		return
	}

	for _, region := range regions {
		if region.Region == config.Region {
			nodes := strings.Split(region.StorageNodes, ",")
			for _, node := range nodes {
				log.Printf("Starting S3 monitoring for node: %s", node)
				go s3monitor.MonitorTenantsS3Stats(node)
				// if(node=="or1"){

				// }
			}
		}
	}
}
