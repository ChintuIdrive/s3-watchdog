package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type aliasConfigV10 struct {
	URL          string `json:"url"`
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	SessionToken string `json:"sessionToken,omitempty"`
	API          string `json:"api"`
	Path         string `json:"path"`
	License      string `json:"license,omitempty"`
	APIKey       string `json:"apiKey,omitempty"`
	Src          string `json:"src,omitempty"`
}

type AdminClient struct {
	client *madmin.AdminClient
}

func NewAdminClient(endpoint, accessKey, secretKey string) (*AdminClient, *probe.Error) {
	aliasCfg := aliasConfigV10{}
	aliasCfg.API = "s3v4"
	aliasCfg.URL = endpoint //e.g:"http://127.0.0.1:9000"
	aliasCfg.AccessKey = accessKey
	aliasCfg.SecretKey = secretKey
	config := NewS3Config("", endpoint, &aliasCfg)

	// Creates a parsed URL.
	targetURL, e := url.Parse(endpoint)
	if e != nil {
		return nil, probe.NewError(e)
	}
	hostName := targetURL.Host

	// Lookup previous cache by hash.
	var api *madmin.AdminClient
	transport := config.getTransport()

	credsChain, err := config.getCredsChain()
	if err != nil {
		return nil, err
	}

	creds := credentials.NewChainCredentials(credsChain)

	// Not found. Instantiate a new MinIO
	api, e = madmin.NewWithOptions(hostName, &madmin.Options{
		Creds:  creds,
		Secure: true,
	})
	if e != nil {
		return nil, probe.NewError(e)
	}

	// Set custom transport.
	api.SetCustomTransport(transport)

	// Set app info.
	api.SetAppInfo(config.AppName, config.AppVersion)

	return &AdminClient{client: api}, nil
}

func (c *AdminClient) AddCannedPolicy(policyName string, policy string) error {
	return c.client.AddCannedPolicy(context.TODO(), policyName, []byte(policy))
}

func (c *AdminClient) AddNewServiceAccount(policy, accessKey, secretKey, name, description string) (madmin.Credentials, error) {
	opts := madmin.AddServiceAccountReq{
		Policy:      []byte(policy),
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Name:        name,
		Description: description,
	}
	return c.client.AddServiceAccount(context.TODO(), opts)
}

func (c *AdminClient) DeleteServiceAccount(accessKey string) error {
	return c.client.DeleteServiceAccount(context.TODO(), accessKey)
}

func (c *AdminClient) AddNotificationCredentials(arns []string, accessKey, secretKey string) (res *http.Response, err error) {
	b, _ := json.Marshal(arns)
	adminPayloadFmt := `{
    "arns": %s,
    "accessKey": "%s",
    "secretKey": "%s"
}`

	adminAccessKey, adminSecretKey := c.client.GetAccessAndSecretKey()
	adminPayload := fmt.Sprintf(adminPayloadFmt, string(b), accessKey, secretKey)
	encryptedData, err := madmin.EncryptData(adminSecretKey, []byte(adminPayload))
	if err != nil {
		return nil, err
	}

	queryValues := url.Values{}
	queryValues.Set("accessKey", adminAccessKey)
	queryValues = url.Values{}
	_ = queryValues

	reqData := madmin.RequestData{}
	reqData.Content = encryptedData
	reqData.RelPath = "/v3/add-notification-credentials"
	//reqData.QueryValues = queryValues

	return c.client.ExecuteMethod(context.TODO(), http.MethodPost, reqData)
}

func (c *AdminClient) RemoveNotificationCredentials(arns []string) (res *http.Response, err error) {
	b, _ := json.Marshal(arns)
	adminPayloadFmt := `{
    "arns": %s
}`

	adminAccessKey, adminSecretKey := c.client.GetAccessAndSecretKey()
	adminPayload := fmt.Sprintf(adminPayloadFmt, string(b))
	encryptedData, err := madmin.EncryptData(adminSecretKey, []byte(adminPayload))
	if err != nil {
		return nil, err
	}

	queryValues := url.Values{}
	queryValues.Set("accessKey", adminAccessKey)
	queryValues = url.Values{}
	_ = queryValues

	reqData := madmin.RequestData{}
	reqData.Content = encryptedData
	reqData.RelPath = "/v3/remove-notification-credentials"

	return c.client.ExecuteMethod(context.TODO(), http.MethodPost, reqData)
}
