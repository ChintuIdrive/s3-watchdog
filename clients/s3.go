package clients

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

// TODO: remove log statements and return proper errors so that
// caller can get sufficient information
type S3Client struct {
	client *s3.Client
}

func NewS3Client(endpoint, accKey, secKey string) *S3Client {
	log.Print("exporting user S3 credentials")
	err := os.Setenv("AWS_ACCESS_KEY_ID", accKey)
	if err != nil {
		log.Print("unable to set S3 ENV AWS_ACCESS_KEY_ID")
		return &S3Client{}
	}

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", secKey)
	if err != nil {
		log.Print("unable to set ENV AWS_SECRET_ACCESS_KEY")
		return &S3Client{}
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(endpoints.UsEast1RegionID),
	)
	if err != nil {
		log.Print(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://" + endpoint)
	})

	return &S3Client{client: client}
}

func NewS3ClientHttp(endpoint, accKey, secKey string) *S3Client {
	log.Print("exporting user S3 credentials", accKey, secKey)
	err := os.Setenv("AWS_ACCESS_KEY_ID", accKey)
	if err != nil {
		log.Print("unable to set S3 ENV AWS_ACCESS_KEY_ID")
		return &S3Client{}
	}

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", secKey)
	if err != nil {
		log.Print("unable to set ENV AWS_SECRET_ACCESS_KEY")
		return &S3Client{}
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(endpoints.UsEast1RegionID),
	)
	if err != nil {
		log.Print(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://" + endpoint)
	})

	return &S3Client{client: client}
}

func (s *S3Client) ListBuckets() ([]types.Bucket, error) {
	var buckets []types.Bucket
	ctx := context.TODO()
	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()
	result, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("Couldn't list buckets for your account. Here's why: %v", err)
		return buckets, err
	}
	if len(result.Buckets) == 0 {
		log.Print("You don't have any buckets!")
	} else {
		buckets = result.Buckets
		sort.Slice(buckets, func(i, j int) bool {
			return buckets[i].CreationDate.After(*buckets[j].CreationDate)
		})
		//log.Printf("listed %d buckets", len(buckets))
	}
	return buckets, err
}
func (client *S3Client) ObjectsCountForBucket(bucketName string) (int, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	result, err := client.client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return 0, err
	}
	return len(result.Contents), nil
}
func (s *S3Client) ListObjectsForBucket(bucket string, numOfPagesToList int) (int, error) {
	if bucket == "" {
		return 0, nil
	}

	params := &s3.ListObjectsV2Input{
		Bucket: &bucket,
	}
	// it will return max 1000 objects
	// result, err := s.client.ListObjectsV2(context.TODO(), params)
	// if err != nil {
	// 	log.Printf("Couldn't list objects in bucket %q. Here's why: %v", bucket, err)
	// }

	// log.Printf("Found %d objects in bucket %q", len(result.Contents), bucket)

	maxKeys := 1000
	paginator := s3.NewListObjectsV2Paginator(s.client, params, func(o *s3.ListObjectsV2PaginatorOptions) {
		if v := int32(maxKeys); v != 0 {
			o.Limit = v
		}
	})

	var objectCount int
	var err error
	if numOfPagesToList >= 1 {
		objectCount, err = getObjCount(paginator, numOfPagesToList)
	} else {
		objectCount, err = getAllObjCount(paginator)
	}
	return objectCount, err
}
func getObjCount(paginator *s3.ListObjectsV2Paginator, numOfPagesToList int) (int, error) {
	var pagesCount int
	var objectCount int
	for paginator.HasMorePages() && (pagesCount < numOfPagesToList) {
		pagesCount++

		// Next Page takes a new context for each page retrieval. This is where
		// you could add timeouts or deadlines.
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return objectCount, err
		}
		//_ = page
		// Log the objects found
		// for _, obj := range page.Contents {
		// 	fmt.Println("Object:", *obj.Key)
		// }
		objectCount = objectCount + len(page.Contents)
		log.Printf("page no:%d objects-count:%d", pagesCount, objectCount)
		//fmt.Println("Num objects:", len(page.Contents))
	}

	log.Println("total pages:", pagesCount)
	log.Println("total objectsCount:", objectCount)

	return objectCount, nil
}
func getAllObjCount(paginator *s3.ListObjectsV2Paginator) (int, error) {
	var pagesCount int
	var objectCount int
	for paginator.HasMorePages() {
		pagesCount++

		// Next Page takes a new context for each page retrieval. This is where
		// you could add timeouts or deadlines.
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return objectCount, err
		}
		//_ = page
		// Log the objects found
		// for _, obj := range page.Contents {
		// 	fmt.Println("Object:", *obj.Key)
		// }
		objectCount = objectCount + len(page.Contents)
		log.Printf("page no:%d objects-count:%d", pagesCount, objectCount)
		//fmt.Println("Num objects:", len(page.Contents))
	}
	log.Println("total pages:", pagesCount)
	log.Println("total objectsCount:", objectCount)
	return objectCount, nil
}

func (s *S3Client) GetObjectsForBucket(bucket string) error {
	params := &s3.ListObjectsV2Input{
		Bucket: &bucket,
	}
	maxKeys := 1000
	paginator := s3.NewListObjectsV2Paginator(s.client, params,
		func(o *s3.ListObjectsV2PaginatorOptions) {
			if v := int32(maxKeys); v != 0 {
				o.Limit = v
			}
		})

	var i int
	for paginator.HasMorePages() {
		i++

		// Next Page takes a new context for each page retrieval. This is where
		// you could add timeouts or deadlines.
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		// Log the objects found
		for _, obj := range page.Contents {
			_, err := s.GetObjectAttributes(bucket, *obj.Key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *S3Client) PutObjects(bucketName string, numObjects int, objectGenerator func(id int) (string, []byte)) error {
	defaultFunc := func(id int) (string, []byte) {
		objName := "object_" + strconv.Itoa(id)
		data := "Stub data for " + objName
		return objName, []byte(data)
	}

	if objectGenerator == nil {
		objectGenerator = defaultFunc
	}

	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	var errState error
	for i := 0; i < numObjects; i++ {
		objName, data := objectGenerator(i)
		input := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objName),
			Body:   bytes.NewReader(data),
			Metadata: map[string]string{
				"metadata1": objName,
				"metadata2": "test_" + strconv.Itoa(i),
			},
		}

		if i%200 == 0 {
			wg.Wait()
			fmt.Println("Done:", i)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.client.PutObject(context.TODO(), input)
			if err != nil {
				lock.Lock()
				defer lock.Unlock()
				errState = err
				log.Println("Error while uploading object:" + objName + ", error:" + err.Error())
			}
		}()
	}
	wg.Wait()
	return errState
}

func (s *S3Client) CreateBucket(bucketName string, enableLocking bool) error {
	fmt.Println("Creating bucket:", bucketName)

	input := &s3.CreateBucketInput{
		Bucket:                     aws.String(bucketName),
		ObjectLockEnabledForBucket: aws.Bool(enableLocking),
	}

	_, err := s.client.CreateBucket(context.TODO(), input)
	return err
}

/*
aws s3api --endpoint-url "https://i5k6.or5.idrivee2-62.com" put-bucket-tagging --bucket mahendra-bucket-tagging --tagging "TagSet=[{Key=Environment,Value=Production},{Key=Owner,Value=YourName}]" --no-verify-ssl
*/

func (s *S3Client) PutBucketTagging(bucketName string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	input := &s3.PutBucketTaggingInput{}
	input.Bucket = aws.String(bucketName)
	input.Tagging = &types.Tagging{}

	input.Tagging.TagSet = make([]types.Tag, len(tags))
	i := 0
	for key, value := range tags {
		input.Tagging.TagSet[i].Key = aws.String(key)
		input.Tagging.TagSet[i].Value = aws.String(value)
		i++
	}

	/*input.Tagging.TagSet = make([]types.Tag, 2)
	input.Tagging.TagSet[0].Key = aws.String("Environment")
	input.Tagging.TagSet[0].Value = aws.String("Production")

	input.Tagging.TagSet[1].Key = aws.String("Owner")
	input.Tagging.TagSet[1].Value = aws.String("VarunSharma")
	fmt.Println(*input)*/

	_, err := s.client.PutBucketTagging(context.TODO(), input)

	return err
}

func (s *S3Client) GetBucketTagging(bucketName string) (*s3.GetBucketTaggingOutput, error) {
	input := &s3.GetBucketTaggingInput{}
	input.Bucket = aws.String(bucketName)
	return s.client.GetBucketTagging(context.TODO(), input)
}

func (s *S3Client) PutBucketLogging(sourceBucket, targetBucket, targetPrefix string) error {
	param := s3.PutBucketLoggingInput{}
	param.Bucket = aws.String(sourceBucket)
	param.BucketLoggingStatus = &types.BucketLoggingStatus{}
	param.BucketLoggingStatus.LoggingEnabled = &types.LoggingEnabled{}
	param.BucketLoggingStatus.LoggingEnabled.TargetBucket = aws.String(targetBucket)
	param.BucketLoggingStatus.LoggingEnabled.TargetPrefix = aws.String(targetPrefix)
	{
		tokf := &types.TargetObjectKeyFormat{}
		tokf.PartitionedPrefix = &types.PartitionedPrefix{}
		tokf.PartitionedPrefix.PartitionDateSource = types.PartitionDateSourceEventTime
		tokf.PartitionedPrefix.PartitionDateSource = types.PartitionDateSourceDeliveryTime
		//tokf.SimplePrefix = &types.SimplePrefix{}
		_ = tokf
		param.BucketLoggingStatus.LoggingEnabled.TargetObjectKeyFormat = tokf
	}
	_, err := s.client.PutBucketLogging(context.TODO(), &param)
	if err != nil {
		fmt.Println("Error in PutBucketLogging:", err)
	}
	return err
}

func (s *S3Client) PutBucketLoggingDisable(sourceBucket string) error {
	param := s3.PutBucketLoggingInput{}
	param.Bucket = aws.String(sourceBucket)
	param.BucketLoggingStatus = &types.BucketLoggingStatus{}
	_, err := s.client.PutBucketLogging(context.TODO(), &param)
	if err != nil {
		fmt.Println("Error in PutBucketLogging:", err)
	}
	return err
}

func (s *S3Client) PutNestedObject(bucketName string, objName string) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objName),
	}

	_, err := s.client.PutObject(context.TODO(), input)
	if err != nil {
		log.Println("Error while uploading object:" + objName + ", error:" + err.Error())
	}
}

func (s *S3Client) GetBucketLogging(sourceBucket string) (*s3.GetBucketLoggingOutput, error) {

	param := s3.GetBucketLoggingInput{}
	param.Bucket = aws.String(sourceBucket)

	output, err := s.client.GetBucketLogging(context.TODO(), &param)
	if err != nil {
		return nil, err
	} else if output == nil {
		return nil, fmt.Errorf("Empty output")
	}
	/*{
		b, _ := json.Marshal(output)

		fmt.Println("Output is:", string(b))
	}*/
	return output, nil
}

func (s *S3Client) DeleteBucket(bucket string) error {
	params := &s3.DeleteBucketInput{}
	params.Bucket = aws.String(bucket)
	_, err := s.client.DeleteBucket(context.TODO(), params)
	return err
}

func (s *S3Client) DeleteObject(bucket, object, versionId string, bypass bool) (*s3.DeleteObjectOutput, error) {
	input := &s3.DeleteObjectInput{}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	if len(versionId) > 0 {
		input.VersionId = aws.String(versionId)
	}
	input.BypassGovernanceRetention = aws.Bool(bypass)
	return s.client.DeleteObject(context.TODO(), input)
}

func (s *S3Client) GetObjectAttributes(bucket, objectKey string) (*s3.GetObjectAttributesOutput, error) {
	params := &s3.GetObjectAttributesInput{}
	params.Bucket = aws.String(bucket)
	params.Key = aws.String(objectKey)

	output, err := s.client.GetObjectAttributes(context.TODO(), params)
	return output, err

	//s.client.GetBucketLocation
}

func (s *S3Client) GetBucketObjectsCount(bucket string) (int, error) {
	if bucket == "" {
		return 0, nil
	}

	params := &s3.ListObjectsV2Input{
		Bucket: &bucket,
	}
	maxKeys := 1000
	paginator := s3.NewListObjectsV2Paginator(s.client, params, func(o *s3.ListObjectsV2PaginatorOptions) {
		if v := int32(maxKeys); v != 0 {
			o.Limit = v
		}
	})

	var i int
	count := 0
	for paginator.HasMorePages() {
		i++

		// Next Page takes a new context for each page retrieval. This is where
		// you could add timeouts or deadlines.
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return 0, err
		}
		count += len(page.Contents)
	}
	return count, nil
}

func (s *S3Client) PutBucketVersioning(bucket string) (*s3.PutBucketVersioningOutput, error) {
	input := s3.PutBucketVersioningInput{}
	input.Bucket = aws.String(bucket)
	input.VersioningConfiguration = &types.VersioningConfiguration{}
	input.VersioningConfiguration.Status = types.BucketVersioningStatusEnabled
	return s.client.PutBucketVersioning(context.TODO(), &input)
}

func (s *S3Client) GetBucketVersioning(bucket string) (*s3.GetBucketVersioningOutput, error) {
	input := s3.GetBucketVersioningInput{}
	input.Bucket = aws.String(bucket)
	return s.client.GetBucketVersioning(context.TODO(), &input)
}

func (s *S3Client) GetObjectVersions(bucket, object string) (*s3.ListObjectVersionsOutput, error) {
	input := &s3.ListObjectVersionsInput{}
	input.Bucket = aws.String(bucket)
	input.Prefix = aws.String(object)

	return s.client.ListObjectVersions(context.TODO(), input)
}

func (s *S3Client) PutBucketGovernanceLock(bucket string, numDays int32) (*s3.PutObjectLockConfigurationOutput, error) {
	input := &s3.PutObjectLockConfigurationInput{}
	input.Bucket = aws.String(bucket)
	input.ObjectLockConfiguration = &types.ObjectLockConfiguration{}
	input.ObjectLockConfiguration.ObjectLockEnabled = types.ObjectLockEnabledEnabled

	rule := &types.ObjectLockRule{}
	rule.DefaultRetention = &types.DefaultRetention{}
	rule.DefaultRetention.Mode = types.ObjectLockRetentionModeGovernance
	rule.DefaultRetention.Days = &numDays

	input.ObjectLockConfiguration.Rule = rule

	return s.client.PutObjectLockConfiguration(context.TODO(), input)
}

func (s *S3Client) PutBucketComplianceLock(bucket string, numDays int32) (*s3.PutObjectLockConfigurationOutput, error) {

	input := &s3.PutObjectLockConfigurationInput{}
	input.Bucket = aws.String(bucket)
	input.ObjectLockConfiguration = &types.ObjectLockConfiguration{}
	input.ObjectLockConfiguration.ObjectLockEnabled = types.ObjectLockEnabledEnabled

	rule := &types.ObjectLockRule{}
	rule.DefaultRetention = &types.DefaultRetention{}
	rule.DefaultRetention.Mode = types.ObjectLockRetentionModeCompliance
	rule.DefaultRetention.Days = &numDays

	input.ObjectLockConfiguration.Rule = rule

	return s.client.PutObjectLockConfiguration(context.TODO(), input)
}

func (s *S3Client) GetBucketLockConfiguration(bucket string) (*s3.GetObjectLockConfigurationOutput, error) {
	input := &s3.GetObjectLockConfigurationInput{}
	input.Bucket = aws.String(bucket)
	output, err := s.client.GetObjectLockConfiguration(context.TODO(), input)
	return output, err
}

func (s *S3Client) PutObjectTag(bucket, object, tagKey, tagValue string) (*s3.PutObjectTaggingOutput, error) {
	input := &s3.PutObjectTaggingInput{}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	input.Tagging = &types.Tagging{}
	input.Tagging.TagSet = make([]types.Tag, 0)
	tag := types.Tag{}
	tag.Key = aws.String(tagKey)
	tag.Value = aws.String(tagValue)
	input.Tagging.TagSet = append(input.Tagging.TagSet, tag)
	return s.client.PutObjectTagging(context.TODO(), input)
}

func (s *S3Client) GetObjectTag(bucket, object string) (*s3.GetObjectTaggingOutput, error) {
	input := &s3.GetObjectTaggingInput{}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	return s.client.GetObjectTagging(context.TODO(), input)
}

func (s *S3Client) DeleteObjectTag(bucket, object string) (*s3.DeleteObjectTaggingOutput, error) {
	input := &s3.DeleteObjectTaggingInput{}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	return s.client.DeleteObjectTagging(context.TODO(), input)
}

func (s *S3Client) PutBucketNotificationConfiguration(bucket string, queueArn, snsArn []string,
	events []types.Event, filtersPrefix, filtersSuffix []string) (*s3.PutBucketNotificationConfigurationOutput, error) {

	filters := &types.NotificationConfigurationFilter{}
	filters.Key = &types.S3KeyFilter{}
	filters.Key.FilterRules = []types.FilterRule{}
	for _, fin := range filtersPrefix {
		filter := types.FilterRule{}
		filter.Name = types.FilterRuleNamePrefix
		filter.Value = aws.String(fin)
		filters.Key.FilterRules = append(filters.Key.FilterRules, filter)
	}

	for _, fin := range filtersSuffix {
		filter := types.FilterRule{}
		filter.Name = types.FilterRuleNameSuffix
		filter.Value = aws.String(fin)
		filters.Key.FilterRules = append(filters.Key.FilterRules, filter)
	}

	input := &s3.PutBucketNotificationConfigurationInput{}
	input.Bucket = aws.String(bucket)
	notificationConfig := &types.NotificationConfiguration{}
	for _, arn := range queueArn {
		queueConfig := types.QueueConfiguration{}
		queueConfig.Events = append(queueConfig.Events, events...)
		queueConfig.QueueArn = aws.String(arn)
		if len(filtersPrefix) > 0 || len(filtersSuffix) > 0 {
			queueConfig.Filter = filters
		}

		notificationConfig.QueueConfigurations = append(notificationConfig.QueueConfigurations, queueConfig)
	}

	for _, arn := range snsArn {
		snsConfig := types.TopicConfiguration{}
		snsConfig.Events = append(snsConfig.Events, events...)
		snsConfig.TopicArn = aws.String(arn)
		if len(filtersPrefix) > 0 || len(filtersSuffix) > 0 {
			snsConfig.Filter = filters
		}

		notificationConfig.TopicConfigurations = append(notificationConfig.TopicConfigurations, snsConfig)
	}

	input.NotificationConfiguration = notificationConfig
	return s.client.PutBucketNotificationConfiguration(context.TODO(), input)
}

func (s *S3Client) GetBucketNotificationConfiguration(bucket string) (*s3.GetBucketNotificationConfigurationOutput, error) {
	input := &s3.GetBucketNotificationConfigurationInput{}
	input.Bucket = aws.String(bucket)
	return s.client.GetBucketNotificationConfiguration(context.TODO(), input)
}

func (s *S3Client) MultipartUpload(bucket, object, filePath string) error {
	input := &s3.CreateMultipartUploadInput{}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	input.ACL = types.ObjectCannedACLPublicRead
	output, err := s.client.CreateMultipartUpload(context.TODO(), input)
	if err != nil {
		return err
	}
	uploadID := *output.UploadId
	fmt.Println("upload id:", uploadID)

	fp, err := os.Open(object)
	if err != nil {
		return err
	}
	fileStat, _ := fp.Stat()
	fileSize := fileStat.Size()
	chunkSize := int64(5244955)
	//chunkSize := int64(5 * 1024 * 1024)
	totalParts := math.Ceil(float64(fileSize) / float64(chunkSize))
	fmt.Println("totalParts:", totalParts)
	partNum := 0
	parts := types.CompletedMultipartUpload{}
	parts.Parts = []types.CompletedPart{}

	//tempFile, _ := os.OpenFile("tempfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	//defer tempFile.Close()

	for true {
		fileChunk := make([]byte, chunkSize)
		n, err := fp.Read(fileChunk)
		if err != nil {
			break
		}

		//tempFile.Write(fileChunk[:n])
		partNum++
		fmt.Println("partNum:", partNum, "Size:", n)
		upInput := &s3.UploadPartInput{}
		upInput.Bucket = aws.String(bucket)
		upInput.Key = aws.String(object)
		upInput.UploadId = aws.String(uploadID)
		upInput.Body = bytes.NewReader(fileChunk[:n])
		{
			tempNum := int32(partNum)
			upInput.PartNumber = &tempNum
		}
		upOutput, err := s.client.UploadPart(context.TODO(), upInput)
		_ = upOutput
		if err != nil {
			fmt.Println("Failed to upload part:", partNum, "error:", err)
			return err
		} else {
			fmt.Println("Successfully uploaded part:", partNum)
		}
		cpPart := types.CompletedPart{}
		cpPart.ETag = upOutput.ETag
		{
			tempNum := int32(partNum)
			cpPart.PartNumber = &tempNum
		}
		parts.Parts = append(parts.Parts, cpPart)
	}
	fp.Close()

	cpInput := &s3.CompleteMultipartUploadInput{}
	cpInput.Bucket = aws.String(bucket)
	cpInput.Key = aws.String(object)
	cpInput.UploadId = aws.String(uploadID)
	cpInput.MultipartUpload = &parts

	_, err = s.client.CompleteMultipartUpload(context.TODO(), cpInput)
	if err != nil {
		fmt.Println("Error in CompleteMultipartUpload, err:", err)
		return err
	}

	return nil
}

/*func SendMessageAnonymous(queueURL string) {
	ctx := context.Background()

	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an SQS client
	client := sqs.New(cfg)

	//	output, err := svc.SendMessage(&sqs.SendMessageInput{
	//		DelaySeconds: aws.Int64(10),
	//		MessageAttributes: map[string]*sqs.MessageAttributeValue{
	//			"Title": &sqs.MessageAttributeValue{
	//				DataType:    aws.String("String"),
	//				StringValue: aws.String("The Whistler"),
	//			},
	//			"Author": &sqs.MessageAttributeValue{
	//				DataType:    aws.String("String"),
	//				StringValue: aws.String("John Grisham"),
	//			},
	//			"WeeksOn": &sqs.MessageAttributeValue{
	//				DataType:    aws.String("Number"),
	//				StringValue: aws.String("6"),
	//			},
	//		},
	//		MessageBody: aws.String("Information about current NY Times fiction bestseller for week of 12/11/2016."),
	//		QueueUrl:    aws.String(queueURL),
	//	})
}*/

//	func SendMessageAnonymous2(queueURL string) {
//		awsCfg, err := config.LoadDefaultConfig(
//			context.Background(),
//			config.WithRegion("eu-north-1"),
//			config.WithCredentialsProvider(aws.AnonymousCredentials{}),
//		)
//
//		sess := session.Must(session.NewSessionWithOptions(session.Options{
//			Config: awsCfg,
//		}))
//		sqsClient := sqs.New(sess) //sqs.NewFromConfig(awsCfg)
//
//		output, err := svc.SendMessage(&sqs.SendMessageInput{
//			DelaySeconds: aws.Int64(10),
//			MessageAttributes: map[string]*sqs.MessageAttributeValue{
//				"Title": &sqs.MessageAttributeValue{
//					DataType:    aws.String("String"),
//					StringValue: aws.String("The Whistler"),
//				},
//				"Author": &sqs.MessageAttributeValue{
//					DataType:    aws.String("String"),
//					StringValue: aws.String("John Grisham"),
//				},
//				"WeeksOn": &sqs.MessageAttributeValue{
//					DataType:    aws.String("Number"),
//					StringValue: aws.String("6"),
//				},
//			},
//			MessageBody: aws.String("Information about current NY Times fiction bestseller for week of 12/11/2016."),
//			QueueUrl:    aws.String(queueURL),
//		})
//		if err != nil {
//			fmt.Println("Error is:", err)
//		} else {
//			fmt.Println("output:", *output)
//		}
//	}
