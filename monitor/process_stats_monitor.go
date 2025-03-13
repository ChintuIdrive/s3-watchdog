package monitor

import (
	"ChintuIdrive/s3-watchdog/actions"
	"ChintuIdrive/s3-watchdog/clients"
	"ChintuIdrive/s3-watchdog/collector"
	"ChintuIdrive/s3-watchdog/conf"
	"ChintuIdrive/s3-watchdog/dto"
	"encoding/json"
	"log"
	"time"
)

type S3StatsMonitor struct {
	config             *conf.Config
	apiServerClient    *clients.APIserverClient
	s3MetricsCollector *collector.S3MetricCollector
	nodeUserMap        map[string][]*collector.S3Metrics
}

func NewS3StatsMonitor(config *conf.Config, s3mc *collector.S3MetricCollector, ac *clients.APIserverClient) *S3StatsMonitor {
	return &S3StatsMonitor{
		config:             config,
		s3MetricsCollector: s3mc,
		apiServerClient:    ac,
		nodeUserMap:        make(map[string][]*collector.S3Metrics),
	}
}

func (ssm *S3StatsMonitor) MonitorTenantsS3Stats(node string) {

	for {
		users, err := ssm.apiServerClient.GetNodeDetails(node)
		if err != nil {
			//notify watchdog not able to fetch tenantlist from api server
			log.Printf("Failed to fetch user list from storage node: %v", err)
			continue
		}
		for _, user := range users {
			ssm.processUser(user, node)
		}
		interval := time.Duration(ssm.config.MonitorInterval) * time.Minute
		log.Printf("%s wait for %s before fetching s3-metrics ", node, interval.String())

		time.Sleep(interval) // Adjust interval as needed

	}

}

func (ssm *S3StatsMonitor) processUser(user dto.User, node string) {
	var s3metric *collector.S3Metrics
	var err error
	var exist bool
	cred, err := ssm.apiServerClient.GetCredential(user)
	if err != nil {
		//notify watchdog not able to fetch tenantlist from api server
		log.Printf("Failed to fetch credential list from storage node: %v", err)
		return
	}

	exist, s3metric = ssm.IsMetricAvailable(node, user.StorageDNS)
	if exist {
		ssm.s3MetricsCollector.UpdateMetricValue(s3metric, cred)
	} else {
		s3metric, err = ssm.s3MetricsCollector.CollectS3Metrics(user.StorageDNS, cred)
		if err != nil {
			//Notify it why it is not able to get the s3 metics
			log.Printf("Failed to collect S3 metrics for tenant %s: %v", user.StorageDNS, err)
			return
		}
		s3metrics := ssm.nodeUserMap[node]
		ssm.nodeUserMap[node] = append(s3metrics, s3metric)
	}

	ssm.checkS3stats(node, user.StorageDNS, s3metric)

}

func (psm *S3StatsMonitor) checkS3stats(node, dns string, s3metric *collector.S3Metrics) {

	notify, msg := s3metric.BucketListingMetric.MonitorThresholdWithDuration()
	if notify {
		psm.NotifyS3Stats(node, dns, msg, s3metric.BucketListingMetric)
	}

	for _, objMetric := range s3metric.ObjectMetricsMap {
		notify, msg := objMetric.ObjecttListingMetric.MonitorThresholdWithDuration()
		if notify {
			psm.NotifyS3Stats(node, dns, msg, objMetric.ObjecttListingMetric)
		}
	}

}

func (psm *S3StatsMonitor) NotifyS3Stats(node, dns, msg string, metric *dto.Metric[time.Duration]) {

	s3Not := actions.S3Notification[time.Duration]{
		Type:      actions.S3Metric,
		NodeId:    node,
		TimeStamp: time.Now().Format(time.RFC3339),
		Metric:    metric,
		Actions:   []actions.Action{actions.Notify},
		Message:   msg,
		S3Dns:     dns,
	}

	payload, err := json.Marshal(s3Not)
	if err != nil {
		log.Printf("Error marshalling system notification: %v", err)
		return
	}

	psm.apiServerClient.Notify(payload)
}

func (ssm *S3StatsMonitor) IsMetricAvailable(node, dns string) (bool, *collector.S3Metrics) {
	s3metrics, exist := ssm.nodeUserMap[node]
	if exist {
		for _, s3metric := range s3metrics {
			if s3metric.DNS == dns {
				return true, s3metric
			}
		}
	}
	return false, nil
}
