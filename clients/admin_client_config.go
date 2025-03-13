// Copyright (c) 2015-2022 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package clients

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/klauspost/compress/gzhttp"
	"github.com/minio/mc/pkg/deadlineconn"
	"github.com/minio/mc/pkg/limiter"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	// Version - version time.RFC3339.
	Version = "DEVELOPMENT.GOGET"
	// ReleaseTag - release tag in TAG.%Y-%m-%dT%H-%M-%SZ.
	ReleaseTag = "DEVELOPMENT.GOGET"
	// CommitID - latest commit id.
	CommitID = "DEVELOPMENT.GOGET"
	// ShortCommitID - first 12 characters from CommitID.
	ShortCommitID = CommitID[:12]
	// CopyrightYear - dynamic value of the copyright end year
	CopyrightYear = "0000"

	globalDebug    = false
	globalInsecure = false
)

func NewS3Config(alias, urlStr string, aliasCfg *aliasConfigV10) *Config {
	// We have a valid alias and hostConfig. We populate the
	// credentials from the match found in the config file.
	s3Config := new(Config)

	s3Config.AppName = filepath.Base(os.Args[0])
	s3Config.AppVersion = ReleaseTag
	/*s3Config.Debug = globalDebug
	s3Config.Insecure = globalInsecure
	s3Config.ConnReadDeadline = globalConnReadDeadline
	s3Config.ConnWriteDeadline = globalConnWriteDeadline
	s3Config.UploadLimit = int64(globalLimitUpload)
	s3Config.DownloadLimit = int64(globalLimitDownload)*/

	s3Config.HostURL = urlStr
	s3Config.Alias = alias
	if aliasCfg != nil {
		s3Config.AccessKey = aliasCfg.AccessKey
		s3Config.SecretKey = aliasCfg.SecretKey
		s3Config.SessionToken = aliasCfg.SessionToken
		s3Config.Signature = aliasCfg.API
		s3Config.Lookup = getLookupType(aliasCfg.Path)
	}
	return s3Config
}

// Config - see http://docs.amazonwebservices.com/AmazonS3/latest/dev/index.html?RESTAuthentication.html
type Config struct {
	Alias             string
	AccessKey         string
	SecretKey         string
	SessionToken      string
	Signature         string
	HostURL           string
	AppName           string
	AppVersion        string
	Debug             bool
	Insecure          bool
	Lookup            minio.BucketLookupType
	ConnReadDeadline  time.Duration
	ConnWriteDeadline time.Duration
	UploadLimit       int64
	DownloadLimit     int64
	Transport         http.RoundTripper
}

type notifyExpiringTLS struct {
	transport http.RoundTripper
}

type ClientURLType int

// ClientURL url client url structure
type ClientURL struct {
	Type            ClientURLType
	Scheme          string
	Host            string
	Path            string
	SchemeSeparator string
	Separator       rune
}

// enum types
const (
	objectStorage = iota // MinIO and S3 compatible cloud storage
	fileSystem           // POSIX compatible file systems
)

// getCredsChain returns an []credentials.Provider array for the config
// and the STS configuration (if present)
func (config *Config) getCredsChain() ([]credentials.Provider, *probe.Error) {
	var credsChain []credentials.Provider
	signType := credentials.SignatureV4
	if strings.EqualFold(config.Signature, "s3v2") {
		signType = credentials.SignatureV2
	}

	// Credentials
	creds := &credentials.Static{
		Value: credentials.Value{
			AccessKeyID:     config.AccessKey,
			SecretAccessKey: config.SecretKey,
			SessionToken:    config.SessionToken,
			SignerType:      signType,
		},
	}
	credsChain = append(credsChain, creds)
	return credsChain, nil
}

// getTransport returns a corresponding *http.Transport for the *Config
// set withS3v2 bool to true to add traceV2 tracer.
func (config *Config) getTransport() http.RoundTripper {
	if config.Transport == nil {
		config.initTransport(true)
	}
	return config.Transport
}

/*func (config *Config) isTLS() bool {
	if stsEndpoint := env.Get("MC_STS_ENDPOINT_"+config.Alias, ""); stsEndpoint != "" {
		stsEndpointURL, err := url.Parse(stsEndpoint)
		if err != nil {
			return false
		}
		return isHostTLS(config) || stsEndpointURL.Scheme == "https"
	}
	return isHostTLS(config)
}*/

func (config *Config) initTransport(withS3v2 bool) {
	var transport http.RoundTripper

	//useTLS := config.isTLS()

	if config.Transport != nil {
		transport = config.Transport
	} else {
		tr := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           newCustomDialContext(config),
			MaxIdleConnsPerHost:   1024,
			WriteBufferSize:       32 << 10, // 32KiB moving up from 4KiB default
			ReadBufferSize:        32 << 10, // 32KiB moving up from 4KiB default
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
			// Set this value so that the underlying transport round-tripper
			// doesn't try to auto decode the body of objects with
			// content-encoding set to `gzip`.
			//
			// Refer:
			//    https://golang.org/src/net/http/transport.go?h=roundTrip#L1843
			DisableCompression: true,
		}
		//if useTLS {
		/*tr.DialTLSContext = newCustomDialTLSContext(&tls.Config{
			RootCAs:            nil, //globalRootCAs,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: config.Insecure,
		})*/

		// Because we create a custom TLSClientConfig, we have to opt-in to HTTP/2.
		// See https://github.com/golang/go/issues/14275
		//
		// TODO: Enable http2.0 when upstream issues related to HTTP/2 are fixed.
		//
		// if e = http2.ConfigureTransport(tr); e != nil {
		// 	return nil, probe.NewError(e)
		// }
		//}
		transport = tr
	}

	transport = limiter.New(config.UploadLimit, config.DownloadLimit, transport)

	/*if config.Debug {
		if strings.EqualFold(config.Signature, "S3v4") {
			transport = httptracer.GetNewTraceTransport(newTraceV4(), transport)
		} else if strings.EqualFold(config.Signature, "S3v2") && withS3v2 {
			transport = httptracer.GetNewTraceTransport(newTraceV2(), transport)
		}
	} else {
		if !globalJSONLine && !globalJSON {
			transport = notifyExpiringTLS{transport: transport}
		}
	}*/

	transport = gzhttp.Transport(transport)
	config.Transport = transport
}

// SelectObjectOpts - opts entered for select API
type SelectObjectOpts struct {
	InputSerOpts    map[string]map[string]string
	OutputSerOpts   map[string]map[string]string
	CompressionType minio.SelectCompressionType
}

func getLookupType(l string) minio.BucketLookupType {
	l = strings.ToLower(l)
	switch l {
	case "off":
		return minio.BucketLookupDNS
	case "on":
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

func isHostTLS(config *Config) bool {
	// By default enable HTTPs.
	useTLS := true
	targetURL := newClientURL(config.HostURL)
	if targetURL.Scheme == "http" {
		useTLS = false
	}
	return useTLS
}

type dialContext func(ctx context.Context, network, addr string) (net.Conn, error)

func newCustomDialContext(c *Config) dialContext {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 15 * time.Second,
		}

		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		dconn := deadlineconn.New(conn).
			WithReadDeadline(c.ConnReadDeadline).
			WithWriteDeadline(c.ConnWriteDeadline)

		return dconn, nil
	}
}

func newClientURL(urlStr string) *ClientURL {
	scheme, rest := getScheme(urlStr)
	if strings.HasPrefix(rest, "//") {
		// if rest has '//' prefix, skip them
		var authority string
		authority, rest = splitSpecial(rest[2:], "/", false)
		if rest == "" {
			rest = "/"
		}
		host := getHost(authority)
		if host != "" && (scheme == "http" || scheme == "https") {
			return &ClientURL{
				Scheme:          scheme,
				Type:            objectStorage,
				Host:            host,
				Path:            rest,
				SchemeSeparator: "://",
				Separator:       '/',
			}
		}
	}
	return &ClientURL{
		Type:      fileSystem,
		Path:      rest,
		Separator: filepath.Separator,
	}
}

func getScheme(rawurl string) (scheme, path string) {
	urlSplits := strings.Split(rawurl, "://")
	if len(urlSplits) == 2 {
		scheme, uri := urlSplits[0], "//"+urlSplits[1]
		// ignore numbers in scheme
		validScheme := regexp.MustCompile("^[a-zA-Z]+$")
		if uri != "" {
			if validScheme.MatchString(scheme) {
				return scheme, uri
			}
		}
	}
	return "", rawurl
}

func splitSpecial(s, delimiter string, cutdelimiter bool) (string, string) {
	i := strings.Index(s, delimiter)
	if i < 0 {
		// if delimiter not found return as is.
		return s, ""
	}
	// if delimiter should be removed, remove it.
	if cutdelimiter {
		return s[0:i], s[i+len(delimiter):]
	}
	// return split strings with delimiter
	return s[0:i], s[i:]
}

func getHost(authority string) (host string) {
	i := strings.LastIndex(authority, "@")
	if i >= 0 {
		// TODO support, username@password style userinfo, useful for ftp support.
		return
	}
	return authority
}
