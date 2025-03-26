package api

import (
	"ChintuIdrive/s3-watchdog/monitor"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	S3StatsMonitor *monitor.S3StatsMonitor
}

func NewHandler(s3StatsMonitor *monitor.S3StatsMonitor) *Handler {
	return &Handler{
		S3StatsMonitor: s3StatsMonitor,
	}
}
func (h *Handler) GetS3Metric(c *gin.Context) {
	node := c.Query("node")
	if node == "" {
		c.JSON(400, gin.H{"error": "node parameter is required"})
		return
	}
	s3Stats := h.S3StatsMonitor.GetS3Metric(node)
	c.JSON(200, s3Stats)
}
