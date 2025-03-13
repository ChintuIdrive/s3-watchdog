package clients

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	SQS_SERVICE = "sqs"
	SNS_SERVICE = "sns"
	S3_SERVICE  = "s3"

	timeFormatV4 = "20060102T150405Z"
)

var necessaryDefaults = map[string]string{
	"Content-Type": "application/x-www-form-urlencoded; charset=utf-8",
	"X-Amz-Date":   timestampV4(),
}

type metadata struct {
	algorithm       string
	credentialScope string
	signedHeaders   string
	date            string
	region          string
	service         string
}

// Credentials stores the information necessary to authorize with AWS and it
// is from this information that requests are signed.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SecurityToken   string `json:"Token"`
	Expiration      time.Time
}

// SignV4 signs a request with Signed Signature Version 4.
func SignV4(request *http.Request, keys Credentials) *http.Request {

	// Add the X-Amz-Security-Token header when using STS
	if keys.SecurityToken != "" {
		request.Header.Set("X-Amz-Security-Token", keys.SecurityToken)
	}

	prepareRequestV4(request)
	meta := new(metadata)

	// Task 1
	hashedCanonReq := hashedCanonicalRequestV4(request, meta)

	// Task 2
	stringToSign := stringToSignV4(request, hashedCanonReq, meta)

	// Task 3
	signingKey := signingKeyV4(keys.SecretAccessKey, meta.date, meta.region, meta.service)
	signature := signatureV4(signingKey, stringToSign)

	//fmt.Println("Meta is:", meta)
	request.Header.Set("Authorization", buildAuthHeaderV4(signature, meta, keys))

	return request
}

func hashedCanonicalRequestV4(request *http.Request, meta *metadata) string {
	// TASK 1. http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
	payload := readAndReplaceBody(request)
	payloadHash := hashSHA256(payload)
	request.Header.Set("X-Amz-Content-Sha256", payloadHash)

	// Set this in header values to make it appear in the range of headers to sign
	request.Header.Set("Host", request.Host)

	var sortedHeaderKeys []string
	for key, _ := range request.Header {
		switch key {
		case "Content-Type", "Content-Md5", "Host":
		default:
			if !strings.HasPrefix(key, "X-Amz-") {
				continue
			}
		}
		sortedHeaderKeys = append(sortedHeaderKeys, strings.ToLower(key))
	}
	sort.Strings(sortedHeaderKeys)

	var headersToSign string
	for _, key := range sortedHeaderKeys {
		value := strings.TrimSpace(request.Header.Get(key))
		if key == "host" {
			//AWS does not include port in signing request.
			if strings.Contains(value, ":") {
				split := strings.Split(value, ":")
				port := split[1]
				if port == "80" || port == "443" {
					value = split[0]
				}
			}
		}
		headersToSign += key + ":" + value + "\n"
	}
	meta.signedHeaders = concat(";", sortedHeaderKeys...)
	canonicalRequest := concat("\n", request.Method, normuri(request.URL.Path), normquery(request.URL.Query()), headersToSign, meta.signedHeaders, payloadHash)

	return hashSHA256([]byte(canonicalRequest))
}

func stringToSignV4(request *http.Request, hashedCanonReq string, meta *metadata) string {
	// TASK 2. http://docs.aws.amazon.com/general/latest/gr/sigv4-create-string-to-sign.html

	requestTs := request.Header.Get("X-Amz-Date")

	meta.algorithm = "AWS4-HMAC-SHA256"
	meta.service, meta.region = serviceAndRegion(request.Host)
	meta.service = "s3"
	meta.date = tsDateV4(requestTs)
	meta.credentialScope = concat("/", meta.date, meta.region, meta.service, "aws4_request")

	return concat("\n", meta.algorithm, requestTs, meta.credentialScope, hashedCanonReq)
}

func signatureV4(signingKey []byte, stringToSign string) string {
	// TASK 3. http://docs.aws.amazon.com/general/latest/gr/sigv4-calculate-signature.html

	return hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
}

func prepareRequestV4(request *http.Request) *http.Request {
	for header, value := range necessaryDefaults {
		if request.Header.Get(header) == "" {
			request.Header.Set(header, value)
		}
	}

	if request.URL.Path == "" {
		request.URL.Path += "/"
	}

	return request
}

func signingKeyV4(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

func buildAuthHeaderV4(signature string, meta *metadata, keys Credentials) string {
	credential := keys.AccessKeyID + "/" + meta.credentialScope

	return meta.algorithm +
		" Credential=" + credential +
		", SignedHeaders=" + meta.signedHeaders +
		", Signature=" + signature
}

func readAndReplaceBody(request *http.Request) []byte {
	if request.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(request.Body)
	request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func timestampV4() string {
	return now().Format(timeFormatV4)
}

func tsDateV4(timestamp string) string {
	return timestamp[:8]
}

/*var (
	awsSignVersion = map[string]int{
		"autoscaling":          4,
		"ce":                   4,
		"cloudfront":           4,
		"cloudformation":       4,
		"cloudsearch":          4,
		"monitoring":           4,
		"dynamodb":             4,
		"ec2":                  4,
		"elasticmapreduce":     4,
		"elastictranscoder":    4,
		"elasticache":          4,
		"es":                   4,
		"glacier":              4,
		"kinesis":              4,
		"redshift":             4,
		"rds":                  4,
		"sdb":                  2,
		"sns":                  4,
		"sqs":                  4,
		"s3":                   4,
		"elasticbeanstalk":     4,
		"importexport":         4,
		"iam":                  4,
		"route53":              4,
		"elasticloadbalancing": 4,
		"email":                4,
	}
)

const (
	envAccessKey       = "AWS_ACCESS_KEY"
	envAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envSecretKey       = "AWS_SECRET_KEY"
	envSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envSecurityToken   = "AWS_SECURITY_TOKEN"
)*/
