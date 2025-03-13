package clients

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"
)

type WebAdminClient struct {
	creds Credentials
	url   string
}

func NewWebAdminClient(accessKey, secretKey, url string) *WebAdminClient {
	cred := Credentials{}
	cred.AccessKeyID = accessKey
	cred.SecretAccessKey = secretKey
	fmt.Println("wac url:", url)
	return &WebAdminClient{creds: cred, url: url}
}

func (w *WebAdminClient) SendMessage(path, payload string) (*http.Response, error) {
	bodyReader := strings.NewReader(payload)

	fmt.Println("url is:", w.url)
	queryPath := filepath.Join("minio/admin/v3", path)
	url := w.url + "/" + queryPath

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req = SignV4(req, w.creds)
	{
		res, err := httputil.DumpRequest(req, true)
		if err == nil {
			fmt.Println("Signed Request is:", string(res))
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
