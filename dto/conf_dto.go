package dto

import (
	"ChintuIdrive/s3-watchdog/cryption"
	"fmt"
	"os/exec"
	"time"
)

type ApiServerConfig struct {
	NodeId        string `json:"node-id"`
	APIPort       string `json:"api-port"`
	APIServerKey  string `json:"api-server-key"`
	APIServerDNS  string `json:"api-server-dns"`
	TenantListApi string `json:"tenant-list-api"`
	Notify        string `json:"notify"`
}

type ControllerConfig struct {
	AccessKeyDir         string `json:"access-keys-dir"`
	ControllerDNS        string `json:"controller-dns"`
	AddServiceAccountApi string `json:"add-service-account-api"`
	GetTenantInfoApi     string `json:"get-tenant-info-api"`
}

type S3Info struct {
	S3Credentials cryption.SecretData `json:"s3-credential"`
	S3Config      S3Config            `json:"s3-config"`
}
type S3Config struct {
	DNS            string `json:"dns"`
	BucketSelector int    `json:"bucket-selector"`
	PageSelector   int    `json:"page-selector"`
}

type Selector struct {
	BucketSelector int `json:"bucket-selector"`
	PageSelector   int `json:"page-selector"`
}
type MetricConfig struct {
	Name     string `json:"name"`
	Limit    int    `json:"limit"`
	Duration int    `json:"duration"`
	//HighLoadDuration time.Duration //Represents how long the value has been above the threshold
}

type MonitoredProcess struct {
	Name             string
	PID              int32
	CPUusage         *Metric[float64]
	MemUsage         *Metric[float64]
	ConnectionsCount *Metric[int]
}

type MonitoredTenantProcess struct {
	*MonitoredProcess
	DNS string `json:"dns"`
}

type TenantS3Metrics struct {
	*MonitoredTenantProcess
	BucketListing       *Metric[time.Duration]
	ObjectListing       *Metric[time.Duration]
	BucketObjectListMap map[string]*Metric[time.Duration]
}

func (mp *MonitoredProcess) Start() {
	// Start the process

	cmd := exec.Command("systemctl", "start", mp.Name)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to start process %s: %v\n", mp.Name, err)
	}
}

func (mp *MonitoredProcess) Stop() {
	// Stop the process

	cmd := exec.Command("systemctl", "stop", mp.Name)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to stop process %s: %v\n", mp.Name, err)
	}
}

func (mp *MonitoredProcess) Restart() {
	// Restart the process

	cmd := exec.Command("systemctl", "restart", mp.Name)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to restart process %s: %v\n", mp.Name, err)
	}
}

func (mp *MonitoredProcess) Notify() {
	// Notify the user

	fmt.Printf("Notify for process %s\n", mp.Name)
}

func (mtp *MonitoredTenantProcess) Start() {
	// Start the process

	cmd := exec.Command("systemctl", "start", mtp.Name)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to start process %s: %v\n", mtp.Name, err)
	}
}

func (mtp *MonitoredTenantProcess) Stop() {
	// Stop the process

	cmd := exec.Command("systemctl", "stop", mtp.Name)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to stop process %s: %v\n", mtp.Name, err)
	}
}

func (mtp *MonitoredTenantProcess) Restart() {
	// Restart the process

	cmd := exec.Command("systemctl", "restart", mtp.Name)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to restart process %s: %v\n", mtp.Name, err)
	}
}
