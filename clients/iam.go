package clients

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

type IAMClient struct {
	client *iam.Client
}

func NewIAMClient(endpoint, accKey, secKey string) *IAMClient {
	err := os.Setenv("AWS_ACCESS_KEY_ID", accKey)
	if err != nil {
		log.Print("unable to set S3 ENV AWS_ACCESS_KEY_ID")
		return nil
	}

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", secKey)
	if err != nil {
		log.Print("unable to set ENV AWS_SECRET_ACCESS_KEY")
		return nil
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(endpoints.UsEast1RegionID),
	)
	if err != nil {
		log.Print(err)
	}

	client := iam.NewFromConfig(cfg, func(o *iam.Options) {
		o.BaseEndpoint = aws.String("http://" + endpoint)
	})

	return &IAMClient{client: client}
}

func (i *IAMClient) CreateIAMPolicy(policyName, policy, description string) (*iam.CreatePolicyOutput, error) {
	//	input := &s3.PutBucketPolicyInput{}
	//	input.Bucket = aws.String(bucket)
	//	input.Policy = aws.String(policy)
	//	output, err := s.client.PutBucketPolicy(context.TODO(), input)
	//	_ = iam.Client
	//	return output, err

	input := &iam.CreatePolicyInput{}
	input.Description = aws.String(description)
	input.PolicyName = aws.String(policyName)
	input.PolicyDocument = aws.String(policy)
	return i.client.CreatePolicy(context.TODO(), input)
}
