package handlers

import (
	"net/http"
	"time"

	"court-scheduler/models"
	"court-scheduler/services"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
)

func ListNotices(c *gin.Context) {
	hearingID := c.Query("hearing_id")
	status := c.Query("status")
	noticeType := c.Query("type")

	notices := store.GetAllNotices()
	result := []models.Notice{}

	for _, n := range notices {
		if hearingID != "" && n.HearingID != hearingID {
			continue
		}
		if status != "" && string(n.Status) != status {
			continue
		}
		if noticeType != "" && string(n.NoticeType) != noticeType {
			continue
		}
		result = append(result, n)
	}

	c.JSON(http.StatusOK, result)
}

func GetNotice(c *gin.Context) {
	id := c.Param("id")
	n, ok := store.GetNotice(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notice not found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

type UpdateNoticeStatusRequest struct {
	Status models.NoticeStatus `json:"status" binding:"required"`
}

func UpdateNoticeStatus(c *gin.Context) {
	id := c.Param("id")
	n, ok := store.GetNotice(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notice not found"})
		return
	}

	var req UpdateNoticeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	n.Status = req.Status
	if req.Status == models.NoticeStatusConfirmed {
		now := time.Now()
		n.ConfirmedAt = &now
	}
	n.UpdatedAt = time.Now()

	store.SaveNotice(n)
	c.JSON(http.StatusOK, n)
}

func GetNoticeTypes(c *gin.Context) {
	result := []map[string]interface{}{}
	for nt, config := range services.NoticeTypeConfig {
		item := map[string]interface{}{
			"type":        nt,
			"name":        config["name"],
			"description": config["description"],
			"channel":     config["channel"],
		}
		result = append(result, item)
	}
	c.JSON(http.StatusOK, result)
}
