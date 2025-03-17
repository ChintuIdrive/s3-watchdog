package conf

import (
	"ChintuIdrive/s3-watchdog/dto"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Region                 string                `json:"region"`
	LogFilePath            string                `json:"log-file-path"`
	ApiServerConfig        *dto.ApiServerConfig  `json:"api-server-config"`
	ControllerConfig       *dto.ControllerConfig `json:"controller-config"`
	BucketListingThreshold *dto.Threshold        `json:"bucket-listing-threshold"`
	ObjectListingThreshold *dto.Threshold        `json:"object-listing-threshold"`
	Selector               *dto.Selector         `json:"selector"`
	MonitorInterval        int                   `json:"monitor-interval"`
	SessionTokenInterval   int                   `json:"session-token-interval"`
	RefreshTokenInterval   int                   `json:"refresh-token-interval"`
	CredentialTTL          int                   `json:"credential-ttl"`
	Login                  dto.Login             `json:"login"`
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func GetDefaultConfig() *Config {
	asc := &dto.ApiServerConfig{
		NodeId:        "nc1",
		APIPort:       ":8080",
		APIServerKey:  "E8AA3FBB0F512B32",
		APIServerDNS:  "e2-api.edgedrive.com",
		TenantListApi: "api/tenant/list",
	}

	return &Config{
		Region: "Charlotte",
		Login: dto.Login{
			Email:     "admin@idrive.com",
			Password:  "ODdkOWJiNDAwYzA2MzQ2OTFmMGUzYmFhZjFlMmZkMGQ=",
			Recaptcha: "sdd",
		},
		LogFilePath: "watchdog.log",

		ApiServerConfig: asc,
		ControllerConfig: &dto.ControllerConfig{
			AccessKeyDir:         "access-keys",
			ControllerDNS:        "localhost:44344",
			AddServiceAccountApi: "admin/v1/add_service_account",
			GetTenantInfoApi:     "admin/v1/get_tenant_info",
		},
		BucketListingThreshold: &dto.Threshold{
			Limit:            5,
			HighLoadDuration: 5,
		},
		ObjectListingThreshold: &dto.Threshold{
			Limit:            5,
			HighLoadDuration: 5,
		},
		Selector: &dto.Selector{
			BucketSelector: 1,
			PageSelector:   1,
		},
		MonitorInterval:      5,
		SessionTokenInterval: 5,
		RefreshTokenInterval: 5,
		CredentialTTL:        5,
	}
}

func (config *Config) AddDefaultS3Config(tenant dto.Tenant) (*dto.S3Config, error) {
	s3config := &dto.S3Config{
		DNS:            tenant.DNS,
		BucketSelector: 1,
		PageSelector:   1,
	}
	s3configDir := filepath.Join(config.ControllerConfig.AccessKeyDir, tenant.DNS)
	if _, err := os.Stat(s3configDir); os.IsNotExist(err) {
		err := os.MkdirAll(s3configDir, os.ModePerm)
		if err != nil {
			//log.Fatalf("Failed to create access key directory: %v", err)
			return nil, err
		}
	}
	s3configPath := filepath.Join(s3configDir, "s3-config.json")
	s3configFile, err := os.Create(s3configPath)
	if err != nil {
		//log.Fatalf("Failed to create s3-config file for tenant %s: %v", tenant.DNS, err)
		return nil, err
	}
	defer s3configFile.Close()

	s3configData, err := json.MarshalIndent(s3config, "", "  ")
	if err != nil {
		//log.Fatalf("Failed to marshal s3-config data for tenant %s: %v", tenant.DNS, err)
		return nil, err
	}

	s3configFile.Write(s3configData)

	//config.TenatS3ConfigMap[tenant.DNS] = s3config

	return s3config, nil
}

func (config *Config) GetS3Config(tenant dto.Tenant) (*dto.S3Config, error) {
	s3configPath := filepath.Join(config.ControllerConfig.AccessKeyDir, tenant.DNS, "s3-config.json")
	data, err := os.ReadFile(s3configPath)
	if err != nil {
		// If the file does not exist, create a default S3Config
		log.Printf("S3 configuration not available for tenant %s, adding default configuration", tenant.DNS)

		return config.AddDefaultS3Config(tenant)
	}

	var s3config dto.S3Config
	err = json.Unmarshal(data, &s3config)
	if err != nil {
		// If there is an error in unmarshalling, return a default S3Config
		return config.AddDefaultS3Config(tenant)
	}

	//config.TenatS3ConfigMap[tenant.DNS] = &s3config
	return &s3config, nil
}

func (config *Config) LoadS3Config(tenantsFromApiServer []dto.Tenant) {
	for _, tenant := range tenantsFromApiServer {
		config.AddDefaultS3Config(tenant)
	}
}
