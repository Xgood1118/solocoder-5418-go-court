package handlers

import (
	"net/http"
	"time"

	"court-scheduler/models"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ListCases(c *gin.Context) {
	cases := store.GetAllCases()
	c.JSON(http.StatusOK, cases)
}

func GetCase(c *gin.Context) {
	id := c.Param("id")
	cs, ok := store.GetCase(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}
	c.JSON(http.StatusOK, cs)
}

type CreateCaseRequest struct {
	CaseNumber    string               `json:"case_number" binding:"required"`
	CaseType      models.CaseType      `json:"case_type" binding:"required"`
	Title         string               `json:"title" binding:"required"`
	Parties       []models.Party       `json:"parties"`
	Lawyers       []models.Lawyer      `json:"lawyers"`
	Witnesses     []models.Witness     `json:"witnesses"`
	CourtroomSize models.CourtroomSize `json:"courtroom_size"`
}

func CreateCase(c *gin.Context) {
	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	cs := models.Case{
		ID:            uuid.New().String(),
		CaseNumber:    req.CaseNumber,
		CaseType:      req.CaseType,
		Title:         req.Title,
		Parties:       req.Parties,
		Lawyers:       req.Lawyers,
		Witnesses:     req.Witnesses,
		CourtroomSize: req.CourtroomSize,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if cs.Parties == nil {
		cs.Parties = []models.Party{}
	}
	if cs.Lawyers == nil {
		cs.Lawyers = []models.Lawyer{}
	}
	if cs.Witnesses == nil {
		cs.Witnesses = []models.Witness{}
	}
	if cs.CourtroomSize == "" {
		switch cs.CaseType {
		case models.CaseTypeCriminal:
			cs.CourtroomSize = models.CourtroomSizeLarge
		default:
			cs.CourtroomSize = models.CourtroomSizeMedium
		}
	}

	store.SaveCase(cs)
	c.JSON(http.StatusCreated, cs)
}

type UpdateCaseRequest struct {
	CaseNumber    *string              `json:"case_number"`
	CaseType      *models.CaseType     `json:"case_type"`
	Title         *string              `json:"title"`
	Parties       *[]models.Party      `json:"parties"`
	Lawyers       *[]models.Lawyer     `json:"lawyers"`
	Witnesses     *[]models.Witness    `json:"witnesses"`
	CourtroomSize *models.CourtroomSize `json:"courtroom_size"`
}

func UpdateCase(c *gin.Context) {
	id := c.Param("id")
	cs, ok := store.GetCase(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	var req UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CaseNumber != nil {
		cs.CaseNumber = *req.CaseNumber
	}
	if req.CaseType != nil {
		cs.CaseType = *req.CaseType
	}
	if req.Title != nil {
		cs.Title = *req.Title
	}
	if req.Parties != nil {
		cs.Parties = *req.Parties
	}
	if req.Lawyers != nil {
		cs.Lawyers = *req.Lawyers
	}
	if req.Witnesses != nil {
		cs.Witnesses = *req.Witnesses
	}
	if req.CourtroomSize != nil {
		cs.CourtroomSize = *req.CourtroomSize
	}

	cs.UpdatedAt = time.Now()
	store.SaveCase(cs)
	c.JSON(http.StatusOK, cs)
}

func DeleteCase(c *gin.Context) {
	id := c.Param("id")
	_, ok := store.GetCase(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}
	store.DeleteCase(id)
	c.JSON(http.StatusNoContent, nil)
}
