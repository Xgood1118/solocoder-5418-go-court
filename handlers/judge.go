package handlers

import (
	"net/http"
	"time"

	"court-scheduler/models"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ListJudges(c *gin.Context) {
	judges := store.GetAllJudges()
	c.JSON(http.StatusOK, judges)
}

func GetJudge(c *gin.Context) {
	id := c.Param("id")
	j, ok := store.GetJudge(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "judge not found"})
		return
	}
	c.JSON(http.StatusOK, j)
}

type CreateJudgeRequest struct {
	Name      string           `json:"name" binding:"required"`
	CaseTypes []models.CaseType `json:"case_types"`
}

func CreateJudge(c *gin.Context) {
	var req CreateJudgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	j := models.Judge{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CaseTypes: req.CaseTypes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if j.CaseTypes == nil {
		j.CaseTypes = []models.CaseType{}
	}

	store.SaveJudge(j)
	c.JSON(http.StatusCreated, j)
}

type UpdateJudgeRequest struct {
	Name      *string            `json:"name"`
	CaseTypes *[]models.CaseType `json:"case_types"`
}

func UpdateJudge(c *gin.Context) {
	id := c.Param("id")
	j, ok := store.GetJudge(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "judge not found"})
		return
	}

	var req UpdateJudgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		j.Name = *req.Name
	}
	if req.CaseTypes != nil {
		j.CaseTypes = *req.CaseTypes
	}

	j.UpdatedAt = time.Now()
	store.SaveJudge(j)
	c.JSON(http.StatusOK, j)
}

func DeleteJudge(c *gin.Context) {
	id := c.Param("id")
	_, ok := store.GetJudge(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "judge not found"})
		return
	}
	store.DeleteJudge(id)
	c.JSON(http.StatusNoContent, nil)
}

func GetJudgeDailyCount(c *gin.Context) {
	id := c.Param("id")
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	hearings := store.GetAllHearings()
	count := 0
	for _, h := range hearings {
		if h.JudgeID == id && h.Date == date && h.Status != models.HearingStatusCancelled && h.Status != models.HearingStatusPostponed {
			count++
		}
	}

	j, ok := store.GetJudge(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "judge not found"})
		return
	}

	maxPerDay := 4
	for _, ct := range j.CaseTypes {
		if ct == models.CaseTypeCriminal {
			maxPerDay = 3
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"judge_id":  id,
		"date":      date,
		"count":     count,
		"max_per_day": maxPerDay,
		"is_overload": count > maxPerDay,
	})
}

func ListJurors(c *gin.Context) {
	jurors := store.GetAllJurors()
	c.JSON(http.StatusOK, jurors)
}

func GetJuror(c *gin.Context) {
	id := c.Param("id")
	j, ok := store.GetJuror(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "juror not found"})
		return
	}
	c.JSON(http.StatusOK, j)
}

type CreateJurorRequest struct {
	Name      string           `json:"name" binding:"required"`
	CaseTypes []models.CaseType `json:"case_types"`
	Phone     string           `json:"phone"`
	Email     string           `json:"email"`
}

func CreateJuror(c *gin.Context) {
	var req CreateJurorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	j := models.Juror{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CaseTypes: req.CaseTypes,
		Phone:     req.Phone,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if j.CaseTypes == nil {
		j.CaseTypes = []models.CaseType{}
	}

	store.SaveJuror(j)
	c.JSON(http.StatusCreated, j)
}

type UpdateJurorRequest struct {
	Name      *string            `json:"name"`
	CaseTypes *[]models.CaseType `json:"case_types"`
	Phone     *string            `json:"phone"`
	Email     *string            `json:"email"`
}

func UpdateJuror(c *gin.Context) {
	id := c.Param("id")
	j, ok := store.GetJuror(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "juror not found"})
		return
	}

	var req UpdateJurorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		j.Name = *req.Name
	}
	if req.CaseTypes != nil {
		j.CaseTypes = *req.CaseTypes
	}
	if req.Phone != nil {
		j.Phone = *req.Phone
	}
	if req.Email != nil {
		j.Email = *req.Email
	}

	j.UpdatedAt = time.Now()
	store.SaveJuror(j)
	c.JSON(http.StatusOK, j)
}

func DeleteJuror(c *gin.Context) {
	id := c.Param("id")
	_, ok := store.GetJuror(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "juror not found"})
		return
	}
	store.DeleteJuror(id)
	c.JSON(http.StatusNoContent, nil)
}
