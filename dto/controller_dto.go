package dto

import "time"

type ServiceAccountReq struct {
	BaseReq
	IsInternal           bool        `json:"isInternal"`
	ValidTillUtc         time.Time   `json:"validTillUtc"`
	Name                 string      `json:"name"`
	Permissions          int         `json:"permissions"`
	Buckets              interface{} `json:"buckets"`
	SecreteId            string      `json:"secreteId"`
	GlobalKeySecret      *string     `json:"globalKeySecret"`
	GlobalKeyId          *string     `json:"globalKeyId"`
	IsTestAccount        bool        `json:"isTestAccount"`
	DisableDeleteVersion bool        `json:"disableDeleteVersion"`
	DisableDeleteBucket  bool        `json:"disableDeleteBucket"`
	DisableDeleteObject  bool        `json:"disableDeleteObject"`
}

type BaseReq struct {
	DNS string    `json:"dns"`
	SID string    `json:"sid"`
	SK  SecretKey `json:"sk"`
}
type TenantProcessInfoReq struct {
	BaseReq
}

type AccessKeys struct {
	AccessKey  string    `json:"accessKey"`
	SecretKey  SecretKey `json:"secretKey"`
	Expiration string    `json:"expiration"`
	StatusCode int       `json:"StatusCode"`
}

type SecretKey struct {
	CString string `json:"CString"`
	DString string `json:"DString"`
}

type TenantProcessInfoResponse struct {
	LocalMinioAdminEndpoint string               `json:"LocalMinioAdminEndpoint"`
	TenatWithProcessInfo    TenatWithProcessInfo `json:"TenantInfo"`
	StatusCode              int                  `json:"StatusCode"`
}
type TenatWithProcessInfo struct {
	PlannedRestart                bool
	PublicBucketsEnabled          bool
	UserType                      string
	E2UserID                      string
	UserID                        string
	DNS                           string
	ProcessID                     int
	Password                      Password
	AdminPort                     int
	S3Port                        int
	FailedS3HealthChecks          int
	ProcessStartTime              string
	CNameList                     []string
	MarkedForForceRestart         bool
	RestartImmediately            bool
	AllowedOrigin                 string
	Compression                   bool
	RestartInProcess              bool
	MaxApiRequests                int
	ApiRequestsDeadline           int
	EnablePublicAccessOnE2URL     bool
	IsFServer                     bool
	TotalRequestsProcessed        int64
	RequestsProcessInLastInterval bool
	Whitelist                     []string
	Blacklist                     []string
	NewServerAdded                bool
	ErrorStat                     int
	UseDEC                        bool
	UseHttp                       bool
	UploadLimit                   int
	DownloadLimit                 int
}
