package collector

import (
	"ChintuIdrive/s3-watchdog/clients"
	"ChintuIdrive/s3-watchdog/conf"
	"ChintuIdrive/s3-watchdog/dto"
	"log"
	"time"
)

type S3Metrics struct {
	DNS                 string                     `json:"dns"`
	BucketsCount        int                        `json:"buckets_count"`
	BucketListingMetric *dto.Metric[time.Duration] `json:"bucket_listing_metric"`
	ObjectMetricsMap    map[string]ObjectMetrics   `json:"object_metrics_map"`
}

type ObjectMetrics struct {
	ObjectsCount         int                        `json:"objects_count"`
	ObjecttListingMetric *dto.Metric[time.Duration] `json:"object_listing_metric"`
}

type S3MetricCollector struct {
	config *conf.Config
}

func NewS3MetricCollector(config *conf.Config) *S3MetricCollector {
	return &S3MetricCollector{
		config: config,
	}
}

func (s3mc *S3MetricCollector) CollectS3Metrics(dns string, cred dto.Cred) (*S3Metrics, error) {

	client := clients.NewS3Client(dns, cred.AccessKey, cred.SecretKey)
	startTime := time.Now()
	buckets, err := client.ListBuckets()
	duration := time.Since(startTime)
	if err != nil {
		return nil, err
	}
	s3metrics := &S3Metrics{
		DNS:          dns,
		BucketsCount: len(buckets),
		BucketListingMetric: &dto.Metric[time.Duration]{
			Name:             "bucket-listing",
			Value:            duration,
			Threshold:        time.Duration(s3mc.config.BucketListingThreshold.Limit) * time.Millisecond,
			HighLoadDuration: s3mc.config.BucketListingThreshold.HighLoadDuration,
		},
		ObjectMetricsMap: make(map[string]ObjectMetrics),
	}

	if s3mc.config.Selector.BucketSelector == 0 {
		log.Printf("No specific bucket selector configured for tenant %s, processing all buckets", dns)
		return s3metrics, nil
	}

	bucketsToProcess := buckets // set s3config.BucketSelector = -1 to process all the buckets
	if s3mc.config.Selector.BucketSelector > 0 && s3mc.config.Selector.BucketSelector < len(buckets) {
		bucketsToProcess = buckets[:s3mc.config.Selector.BucketSelector]
	}

	for _, bucket := range bucketsToProcess {
		startTime = time.Now()
		objCount, err := client.ListObjectsForBucket(*bucket.Name, s3mc.config.Selector.PageSelector)
		duration = time.Since(startTime)
		if err != nil {
			log.Printf("Error listing objects for bucket %s: %v", *bucket.Name, err)
			continue
		}
		objMetric := ObjectMetrics{
			ObjectsCount: objCount,
			ObjecttListingMetric: &dto.Metric[time.Duration]{
				Name:             "object-listing",
				Value:            duration,
				Threshold:        time.Duration(s3mc.config.ObjectListingThreshold.Limit) * time.Millisecond,
				HighLoadDuration: s3mc.config.ObjectListingThreshold.HighLoadDuration,
				LastAlertTime:    time.Time{},
			},
		}
		s3metrics.ObjectMetricsMap[*bucket.Name] = objMetric
		log.Print(*bucket.Name)
	}

	return s3metrics, nil
}

func (s3mc *S3MetricCollector) UpdateMetricValue(metric *S3Metrics, cred dto.Cred) error {
	client := clients.NewS3Client(metric.DNS, cred.AccessKey, cred.SecretKey)
	startTime := time.Now()
	buckets, err := client.ListBuckets()
	duration := time.Since(startTime)
	if err != nil {
		return err
	}
	metric.BucketListingMetric.Value = duration
	for _, bucket := range buckets {
		startTime = time.Now()
		objCount, err := client.ListObjectsForBucket(*bucket.Name, s3mc.config.Selector.PageSelector)
		duration = time.Since(startTime)
		if err != nil {
			log.Printf("Error listing objects for bucket %s: %v", *bucket.Name, err)
			continue
		}
		objmetric, exist := metric.ObjectMetricsMap[*bucket.Name]
		objmetric.ObjectsCount = objCount
		if exist {
			objmetric.ObjecttListingMetric.Value = duration
		} else {
			objMetric := ObjectMetrics{
				ObjectsCount: objCount,
				ObjecttListingMetric: &dto.Metric[time.Duration]{
					Name:             "object-listing",
					Value:            duration,
					Threshold:        time.Duration(s3mc.config.ObjectListingThreshold.Limit) * time.Millisecond,
					HighLoadDuration: s3mc.config.ObjectListingThreshold.HighLoadDuration,
					LastAlertTime:    time.Time{},
				},
			}
			metric.ObjectMetricsMap[*bucket.Name] = objMetric
		}
	}
	return nil
}
