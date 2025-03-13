package dto

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/exp/constraints"
)

type Metric[T constraints.Ordered] struct {
	Name             string        `json:"name"`
	Value            T             `json:"value"`              // Current value of the metric
	Threshold        T             `json:"threshold"`          // Maximum allowed value before an alert
	HighLoadDuration time.Duration `json:"high_load_duration"` //Represents how long the value has been above the threshold
	LastAlertTime    time.Time     `json:"-"`                  // Time of the last alert
	//DNS              string        `json:"dsn"`
}

type SystemMetric[T constraints.Ordered] struct {
	Metric[T]
}

type ProcessMetric[T constraints.Ordered] struct {
	Metric[T]
	PName string `json:"pname"`
	PID   int32  `json:"pid"`
}

type S3Metric[T constraints.Ordered] struct {
	Metric[T]
}

// type Threshold[T any] struct {
// 	Limit    T             `json:"limit"`    // Maximum allowed value before an alert
// 	Duration time.Duration `json:"duration"` // Time period for which the threshold must be exceeded
// }

var (
	alertLock sync.Mutex
)

func NewMetric[T constraints.Ordered](name string, threshold T, highLoadDuration time.Duration) *Metric[T] {
	return &Metric[T]{
		Name:             name,
		Threshold:        threshold,
		HighLoadDuration: highLoadDuration,
	}
}

func (m *Metric[T]) MonitorThresholdWithDuration() (bool, string) {
	alertLock.Lock()
	defer alertLock.Unlock()

	now := time.Now()
	if m.Value > m.Threshold {
		if m.LastAlertTime.IsZero() {
			m.LastAlertTime = now // Start tracking high metric usage time
		} else if now.Sub(m.LastAlertTime) >= m.HighLoadDuration {
			alertMessage := fmt.Sprintf("[ALERT] High %s: %v for %v!", m.Name, m.Value, m.HighLoadDuration)
			log.Println(alertMessage)
			m.LastAlertTime = time.Time{} // Reset timer after alert
			return true, alertMessage
		}
	} else {
		m.LastAlertTime = time.Time{} // Reset if metric usage is normal
	}
	return false, ""
}

func (m *Metric[T]) MonitorImmediateThreshold(disk string) (bool, string) {
	alertLock.Lock()
	defer alertLock.Unlock()

	now := time.Now()
	if m.Value > m.Threshold && now.Sub(m.LastAlertTime) > m.HighLoadDuration {
		alertMessage := fmt.Sprintf("[ALERT] High %s on disk %s: %v", m.Name, disk, m.Value)
		log.Println(alertMessage)
		m.LastAlertTime = now
		return true, alertMessage
	}
	return false, ""
}

// func (m *Metric[T]) executeActions() {
// 	for _, action := range m.Actions {
// 		action()
// 	}
// }

// func (m *Metric[T]) AddAction(actionType ActionType) {
// 	switch actionType {
// 	case NotifyAction:
// 		m.Actions = append(m.Actions, m.notify)
// 	case StartAction:
// 		if m.Type != SystemMetric {
// 			m.Actions = append(m.Actions, m.start)
// 		}
// 	case StopAction:
// 		if m.Type != SystemMetric {
// 			m.Actions = append(m.Actions, m.stop)
// 		}
// 	case RestartAction:
// 		if m.Type != SystemMetric {
// 			m.Actions = append(m.Actions, m.restart)
// 		}
// 	}
// }

func (m *SystemMetric[T]) notify() {
	log.Printf("[ACTION] Notify for %s", m.Name)
	// Call the hook/API provided by the API server
}

func (m *ProcessMetric[T]) notify() {
	log.Printf("[ACTION] Notify for %s", m.Name)
	// Call the hook/API provided by the API server
}

func (m *Metric[T]) start() {
	log.Printf("[ACTION] Start for %s", m.Name)
	// Start tenant/process
}

func (m *Metric[T]) stop() {
	log.Printf("[ACTION] Stop for %s", m.Name)
	// Stop tenant/process
}

func (m *Metric[T]) restart() {
	log.Printf("[ACTION] Restart for %s", m.Name)
	// Restart tenant/process
}

type DiskMetrics struct {
	Name      string            `json:"name"`
	DiskUsage *Metric[float64]  `json:"disk_usage"`
	IoMetrics []*Metric[uint64] `json:"io_metrics"`
}

type SystemMetrics struct {
	ResourceMetrics []*Metric[float64] `json:"resource_metrics"`
	DiskMetrics     []*DiskMetrics     `json:"disk_metrics"`
}

// type SystemLevelThreshold struct {
// 	HighAvgLoad     Threshold[float64] `json:"high_avg_load"`
// 	HighCPUusage    Threshold[float64] `json:"high_cpu_usage"`
// 	HighMemoryUsage Threshold[float64] `json:"high_memory_usage"` // Alert if RAM usage > 90%
// 	HighDiskUsage   Threshold[uint64]  `json:"high_disk_usage"`
// }

// type MemoryUsageThreshold[T any] struct {
// 	// Alert if CPU > 80%
// 	HighMemoryUsage Threshold[T] `json:"high_memory_usage"` // Alert if RAM usage > 90%
// 	HighDiskUsage   Threshold[T] `json:"high_disk_usage"`
// }

type Threshold struct {
	Limit            int `json:"limit"`
	HighLoadDuration int `json:"high_load_duration"`
}
