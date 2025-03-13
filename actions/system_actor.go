package actions

import (
	"ChintuIdrive/s3-watchdog/dto"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"golang.org/x/exp/constraints"
)

type MonitorType string

const (
	SystemMetric  MonitorType = "system"
	ProcessMetric MonitorType = "process"
	S3Metric      MonitorType = "s3"
)

// Action defines the type of action to be taken when a metric reaches its limit
type Action string

const (
	Notify  Action = "notify"
	StartA  Action = "start"
	Stop    Action = "stop"
	Restart Action = "restart"
)

// type SystemNotification[T constraints.Ordered] struct {
// 	Type      MonitorType    `json:"monitor-type"`
// 	NodeId    string         `json:"node-id"`
// 	TimeStamp string         `json:"time-stamp"`
// 	Metric    *dto.Metric[T] `json:"metric"`
// 	Actions   []Action       `json:"actions"`
// 	Message   string         `json:"message"`
// }

// type ProcessNotification[T constraints.Ordered] struct {
// 	SystemNotification[T]
// 	ProcessName string `json:"process-name"`
// 	ProcessId   int32  `json:"process-id"`
// }

type S3Notification[T constraints.Ordered] struct {
	Type      MonitorType    `json:"monitor-type"`
	NodeId    string         `json:"node-id"`
	TimeStamp string         `json:"time-stamp"`
	Metric    *dto.Metric[T] `json:"metric"`
	Actions   []Action       `json:"actions"`
	Message   string         `json:"message"`
	S3Dns     string         `json:"s3-dns"`
}

// Actor is a struct that represents an actor
type Actor struct {
	config *config.Config
}

func (a *Actor) Notify(metric dto.Metric[float64]) {
	log.Printf("[ACTION] Notify for %s", metric.Name)

	// sysNotification:= SystemNotification{
	// 	NodeId:    "node-1",
	// 	TimeStamp: "2021-09-01T12:00:00Z",
	// 	Actions:   []Action{Notify},
	// 	Mesage:    "System metric reached the limit",
	// }
}

func (a *Actor) Start(metric dto.Metric[float64]) {
	log.Printf("[ACTION] Start for %s", metric.Name)
}
