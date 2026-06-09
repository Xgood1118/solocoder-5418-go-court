package handlers

import (
	"net/http"
	"time"

	"court-scheduler/models"
	"court-scheduler/services"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ListHearings(c *gin.Context) {
	date := c.Query("date")
	judgeID := c.Query("judge_id")
	caseID := c.Query("case_id")
	status := c.Query("status")

	hearings := store.GetAllHearings()
	result := []models.Hearing{}

	for _, h := range hearings {
		if date != "" && h.Date != date {
			continue
		}
		if judgeID != "" && h.JudgeID != judgeID {
			continue
		}
		if caseID != "" && h.CaseID != caseID {
			continue
		}
		if status != "" && string(h.Status) != status {
			continue
		}
		result = append(result, h)
	}

	c.JSON(http.StatusOK, result)
}

func GetHearing(c *gin.Context) {
	id := c.Param("id")
	h, ok := store.GetHearing(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "hearing not found"})
		return
	}
	c.JSON(http.StatusOK, h)
}

type CreateHearingRequest struct {
	CaseID      string           `json:"case_id" binding:"required"`
	JudgeID     string           `json:"judge_id" binding:"required"`
	CourtroomID string           `json:"courtroom_id" binding:"required"`
	Date        string           `json:"date" binding:"required"`
	TimeSlot    models.TimeSlot  `json:"time_slot" binding:"required"`
	DurationMin int              `json:"duration_min"`
	JurorCount  int              `json:"juror_count"`
	Translator  string           `json:"translator"`
	Expert      string           `json:"expert"`
}

func CreateHearing(c *gin.Context) {
	var req CreateHearingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store.GlobalStore.ScheduleMutex.Lock()
	defer store.GlobalStore.ScheduleMutex.Unlock()

	hearingCase, ok := store.GetCase(req.CaseID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "case not found"})
		return
	}

	_, ok = store.GetJudge(req.JudgeID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "judge not found"})
		return
	}

	courtroom, ok := store.GetCourtroom(req.CourtroomID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "courtroom not found"})
		return
	}

	if hearingCase.CaseType == models.CaseTypeCriminal && !courtroom.Equipment.DetentionAccess {
		c.JSON(http.StatusBadRequest, gin.H{"error": "courtroom does not have detention access for criminal case"})
		return
	}

	if services.CheckCourtroomOccupied(req.CourtroomID, req.Date, req.TimeSlot, "") {
		c.JSON(http.StatusConflict, models.ScheduleResult{
			Success: false,
			Conflicts: []models.ConflictInfo{
				{Type: "courtroom", Message: "该法庭此时段已被占用"},
			},
		})
		return
	}

	allConflicts := []models.ConflictInfo{}

	judgeConflicts := services.CheckJudgeConflict(req.JudgeID, req.Date, req.TimeSlot, "")
	allConflicts = append(allConflicts, judgeConflicts...)

	lawyerIDs := []string{}
	for _, l := range hearingCase.Lawyers {
		lawyerIDs = append(lawyerIDs, l.ID)
	}
	lawyerConflicts := services.CheckLawyerConflict(lawyerIDs, req.Date, req.TimeSlot, "")
	allConflicts = append(allConflicts, lawyerConflicts...)

	witnessConflicts := services.CheckWitnessConflict(req.CaseID, req.Date, req.TimeSlot, "")
	allConflicts = append(allConflicts, witnessConflicts...)

	if len(allConflicts) > 0 {
		c.JSON(http.StatusConflict, models.ScheduleResult{
			Success:   false,
			Conflicts: allConflicts,
		})
		return
	}

	warnings := []models.ScheduleWarning{}

	dailyCount := services.GetJudgeDailyHearingCount(req.JudgeID, req.Date, "")
	maxPerDay := services.GetMaxHearingsPerDay(hearingCase.CaseType)
	if dailyCount >= maxPerDay {
		warnings = append(warnings, models.ScheduleWarning{
			Level:   "yellow",
			Message: "法官当日庭数已达上限，建议确认是否继续",
		})
	}

	jurorIDs := []string{}
	jurorAudit := models.JurorDrawAudit{}
	if req.JurorCount > 0 {
		ids, audit, err := services.DrawJurors(hearingCase.CaseType, req.Date, req.JurorCount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		jurorIDs = ids
		jurorAudit = audit
	}

	startTime, endTime, err := services.GetSlotTimeRange(req.TimeSlot, req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time slot"})
		return
	}

	if req.DurationMin <= 0 {
		req.DurationMin = int(endTime.Sub(startTime).Minutes())
	}

	now := time.Now()
	hearing := models.Hearing{
		ID:             uuid.New().String(),
		CaseID:         req.CaseID,
		JudgeID:        req.JudgeID,
		CourtroomID:    req.CourtroomID,
		Date:           req.Date,
		TimeSlot:       req.TimeSlot,
		StartTime:      startTime,
		EndTime:        endTime,
		DurationMin:    req.DurationMin,
		Status:         models.HearingStatusScheduled,
		JurorIDs:       jurorIDs,
		JurorCount:     req.JurorCount,
		JurorDrawAudit: jurorAudit,
		Translator:     req.Translator,
		Expert:         req.Expert,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	store.SaveHearing(hearing)

	go services.DispatchNotices(hearing.ID)

	c.JSON(http.StatusCreated, models.ScheduleResult{
		Success:  true,
		Hearing:  &hearing,
		Warnings: warnings,
	})
}

type PostponeHearingRequest struct {
	NewDate     string          `json:"new_date" binding:"required"`
	NewTimeSlot models.TimeSlot `json:"new_time_slot" binding:"required"`
	Reason      string          `json:"reason" binding:"required"`
}

func PostponeHearing(c *gin.Context) {
	id := c.Param("id")

	store.GlobalStore.ScheduleMutex.Lock()
	defer store.GlobalStore.ScheduleMutex.Unlock()

	hearing, ok := store.GetHearing(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "hearing not found"})
		return
	}

	var req PostponeHearingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hearingCase, _ := store.GetCase(hearing.CaseID)

	if services.CheckCourtroomOccupied(hearing.CourtroomID, req.NewDate, req.NewTimeSlot, id) {
		c.JSON(http.StatusConflict, models.ScheduleResult{
			Success: false,
			Conflicts: []models.ConflictInfo{
				{Type: "courtroom", Message: "该法庭新时段已被占用"},
			},
		})
		return
	}

	allConflicts := []models.ConflictInfo{}

	judgeConflicts := services.CheckJudgeConflict(hearing.JudgeID, req.NewDate, req.NewTimeSlot, id)
	allConflicts = append(allConflicts, judgeConflicts...)

	lawyerIDs := []string{}
	for _, l := range hearingCase.Lawyers {
		lawyerIDs = append(lawyerIDs, l.ID)
	}
	lawyerConflicts := services.CheckLawyerConflict(lawyerIDs, req.NewDate, req.NewTimeSlot, id)
	allConflicts = append(allConflicts, lawyerConflicts...)

	witnessConflicts := services.CheckWitnessConflict(hearing.CaseID, req.NewDate, req.NewTimeSlot, id)
	allConflicts = append(allConflicts, witnessConflicts...)

	if len(allConflicts) > 0 {
		c.JSON(http.StatusConflict, models.ScheduleResult{
			Success:   false,
			Conflicts: allConflicts,
		})
		return
	}

	warnings := []models.ScheduleWarning{}

	newDate, _ := time.Parse("2006-01-02", req.NewDate)
	daysUntil := int(time.Until(newDate).Hours() / 24)
	if daysUntil < 3 {
		warnings = append(warnings, models.ScheduleWarning{
			Level:   "red",
			Message: "距新开庭日期不足3天，延期通知可能无法及时送达",
		})
	}

	dailyCount := services.GetJudgeDailyHearingCount(hearing.JudgeID, req.NewDate, id)
	maxPerDay := services.GetMaxHearingsPerDay(hearingCase.CaseType)
	if dailyCount >= maxPerDay {
		warnings = append(warnings, models.ScheduleWarning{
			Level:   "yellow",
			Message: "法官当日庭数已达上限，建议确认是否继续",
		})
	}

	originalDate := hearing.Date
	originalSlot := hearing.TimeSlot
	if hearing.OriginalDate != "" {
		originalDate = hearing.OriginalDate
		originalSlot = hearing.OriginalSlot
	}

	newStart, newEnd, _ := services.GetSlotTimeRange(req.NewTimeSlot, req.NewDate)

	hearing.OriginalDate = originalDate
	hearing.OriginalSlot = originalSlot
	hearing.Date = req.NewDate
	hearing.TimeSlot = req.NewTimeSlot
	hearing.StartTime = newStart
	hearing.EndTime = newEnd
	hearing.Status = models.HearingStatusScheduled
	hearing.PostponeReason = req.Reason
	hearing.UpdatedAt = time.Now()

	store.SaveHearing(hearing)

	go services.DispatchPostponeNotices(hearing.ID)

	c.JSON(http.StatusOK, models.ScheduleResult{
		Success:  true,
		Hearing:  &hearing,
		Warnings: warnings,
	})
}

type CancelHearingRequest struct {
	Reason string `json:"reason"`
}

func CancelHearing(c *gin.Context) {
	id := c.Param("id")
	h, ok := store.GetHearing(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "hearing not found"})
		return
	}

	var req CancelHearingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Status = models.HearingStatusCancelled
	h.PostponeReason = req.Reason
	h.UpdatedAt = time.Now()

	store.SaveHearing(h)
	c.JSON(http.StatusOK, h)
}

func CompleteHearing(c *gin.Context) {
	id := c.Param("id")
	h, ok := store.GetHearing(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "hearing not found"})
		return
	}

	h.Status = models.HearingStatusCompleted
	h.UpdatedAt = time.Now()

	store.SaveHearing(h)
	c.JSON(http.StatusOK, h)
}
