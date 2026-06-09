package handlers

import (
	"net/http"
	"time"

	"court-scheduler/models"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ListCourtrooms(c *gin.Context) {
	courtrooms := store.GetAllCourtrooms()
	c.JSON(http.StatusOK, courtrooms)
}

func GetCourtroom(c *gin.Context) {
	id := c.Param("id")
	cr, ok := store.GetCourtroom(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "courtroom not found"})
		return
	}
	c.JSON(http.StatusOK, cr)
}

type CreateCourtroomRequest struct {
	Name      string                  `json:"name" binding:"required"`
	Size      models.CourtroomSize    `json:"size" binding:"required"`
	Capacity  int                     `json:"capacity"`
	Equipment models.CourtroomEquipment `json:"equipment"`
}

func CreateCourtroom(c *gin.Context) {
	var req CreateCourtroomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	cr := models.Courtroom{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Size:      req.Size,
		Capacity:  req.Capacity,
		Equipment: req.Equipment,
		CreatedAt: now,
		UpdatedAt: now,
	}

	switch req.Size {
	case models.CourtroomSizeLarge:
		if cr.Capacity == 0 {
			cr.Capacity = 50
		}
	case models.CourtroomSizeMedium:
		if cr.Capacity == 0 {
			cr.Capacity = 20
		}
	case models.CourtroomSizeSmall:
		if cr.Capacity == 0 {
			cr.Capacity = 8
		}
	}

	store.SaveCourtroom(cr)
	c.JSON(http.StatusCreated, cr)
}

type UpdateCourtroomRequest struct {
	Name      *string                 `json:"name"`
	Size      *models.CourtroomSize   `json:"size"`
	Capacity  *int                    `json:"capacity"`
	Equipment *models.CourtroomEquipment `json:"equipment"`
}

func UpdateCourtroom(c *gin.Context) {
	id := c.Param("id")
	cr, ok := store.GetCourtroom(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "courtroom not found"})
		return
	}

	var req UpdateCourtroomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		cr.Name = *req.Name
	}
	if req.Size != nil {
		cr.Size = *req.Size
	}
	if req.Capacity != nil {
		cr.Capacity = *req.Capacity
	}
	if req.Equipment != nil {
		cr.Equipment = *req.Equipment
	}

	cr.UpdatedAt = time.Now()
	store.SaveCourtroom(cr)
	c.JSON(http.StatusOK, cr)
}

func DeleteCourtroom(c *gin.Context) {
	id := c.Param("id")
	_, ok := store.GetCourtroom(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "courtroom not found"})
		return
	}
	store.DeleteCourtroom(id)
	c.JSON(http.StatusNoContent, nil)
}
