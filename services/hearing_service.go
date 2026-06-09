package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"court-scheduler/models"
	"court-scheduler/store"
)

func DrawJurors(caseType models.CaseType, date string, count int) ([]string, models.JurorDrawAudit, error) {
	if count < 1 || count > 3 {
		count = 1
	}

	allJurors := store.GetAllJurors()
	candidates := []string{}
	candidateIDs := []string{}

	for _, j := range allJurors {
		matches := false
		for _, ct := range j.CaseTypes {
			if ct == caseType {
				matches = true
				break
			}
		}
		if !matches {
			continue
		}

		occupied := false
		hearings := store.GetAllHearings()
		for _, h := range hearings {
			if h.Date == date && h.Status != models.HearingStatusCancelled && h.Status != models.HearingStatusPostponed {
				for _, jid := range h.JurorIDs {
					if jid == j.ID {
						occupied = true
						break
					}
				}
				if occupied {
					break
				}
			}
		}

		if !occupied {
			candidates = append(candidates, j.ID)
			candidateIDs = append(candidateIDs, j.ID)
		}
	}

	if len(candidates) < count {
		return nil, models.JurorDrawAudit{}, fmt.Errorf("not enough available jurors: need %d, have %d", count, len(candidates))
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := []string{}
	perm := r.Perm(len(candidates))
	for i := 0; i < count; i++ {
		selected = append(selected, candidates[perm[i]])
	}

	audit := models.JurorDrawAudit{
		DrawTime:     time.Now(),
		CandidateIDs: candidateIDs,
		SelectedIDs:  selected,
		Reason:       fmt.Sprintf("random draw for %s case, %d jurors requested", caseType, count),
	}

	return selected, audit, nil
}

func GetSlotTimeRange(slot models.TimeSlot, dateStr string) (time.Time, time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	year, month, day := date.Date()

	switch slot {
	case models.SlotMorning:
		start := time.Date(year, month, day, 9, 0, 0, 0, time.Local)
		end := time.Date(year, month, day, 11, 30, 0, 0, time.Local)
		return start, end, nil
	case models.SlotAfternoon:
		start := time.Date(year, month, day, 14, 0, 0, 0, time.Local)
		end := time.Date(year, month, day, 17, 0, 0, 0, time.Local)
		return start, end, nil
	case models.SlotEvening:
		start := time.Date(year, month, day, 18, 0, 0, 0, time.Local)
		end := time.Date(year, month, day, 20, 0, 0, 0, time.Local)
		return start, end, nil
	case models.SlotWeekendAM:
		start := time.Date(year, month, day, 9, 0, 0, 0, time.Local)
		end := time.Date(year, month, day, 11, 0, 0, 0, time.Local)
		return start, end, nil
	default:
		return time.Time{}, time.Time{}, errors.New("invalid time slot")
	}
}

func CheckJudgeConflict(judgeID string, date string, slot models.TimeSlot, excludeHearingID string) []models.ConflictInfo {
	var conflicts []models.ConflictInfo
	hearings := store.GetAllHearings()

	start, end, err := GetSlotTimeRange(slot, date)
	if err != nil {
		return []models.ConflictInfo{{Type: "error", Message: "invalid time slot"}}
	}

	for _, h := range hearings {
		if h.ID == excludeHearingID {
			continue
		}
		if h.JudgeID != judgeID {
			continue
		}
		if h.Status == models.HearingStatusCancelled || h.Status == models.HearingStatusPostponed {
			continue
		}

		hStart, hEnd, _ := GetSlotTimeRange(h.TimeSlot, h.Date)
		if timesOverlap(start, end, hStart, hEnd) {
			conflicts = append(conflicts, models.ConflictInfo{
				Type:    "judge",
				Message: fmt.Sprintf("法官该时段已有庭期安排"),
				Detail:  fmt.Sprintf("与庭期 %s 冲突，时间: %s %s", h.ID, h.Date, h.TimeSlot),
			})
		}
	}

	return conflicts
}

func CheckLawyerConflict(lawyerIDs []string, date string, slot models.TimeSlot, excludeHearingID string) []models.ConflictInfo {
	var conflicts []models.ConflictInfo
	hearings := store.GetAllHearings()

	start, end, err := GetSlotTimeRange(slot, date)
	if err != nil {
		return []models.ConflictInfo{{Type: "error", Message: "invalid time slot"}}
	}

	hearingLawyerMap := make(map[string][]string)
	for _, h := range hearings {
		if h.ID == excludeHearingID {
			continue
		}
		if h.Status == models.HearingStatusCancelled || h.Status == models.HearingStatusPostponed {
			continue
		}
		cs, ok := store.GetCase(h.CaseID)
		if !ok {
			continue
		}
		for _, l := range cs.Lawyers {
			hearingLawyerMap[h.ID] = append(hearingLawyerMap[h.ID], l.ID)
		}
	}

	for _, lid := range lawyerIDs {
		for hid, lids := range hearingLawyerMap {
			for _, existingLID := range lids {
				if existingLID == lid {
					h, _ := store.GetHearing(hid)
					hStart, hEnd, _ := GetSlotTimeRange(h.TimeSlot, h.Date)
					if timesOverlap(start, end, hStart, hEnd) {
						conflicts = append(conflicts, models.ConflictInfo{
							Type:    "lawyer",
							Message: fmt.Sprintf("律师 %s 该时段已有其他庭期", lid),
							Detail:  fmt.Sprintf("与庭期 %s 冲突，时间: %s %s", hid, h.Date, h.TimeSlot),
						})
					}
				}
			}
		}
	}

	return conflicts
}

func CheckWitnessConflict(caseID string, date string, slot models.TimeSlot, excludeHearingID string) []models.ConflictInfo {
	var conflicts []models.ConflictInfo

	currentCase, ok := store.GetCase(caseID)
	if !ok {
		return []models.ConflictInfo{{Type: "error", Message: "case not found"}}
	}

	if currentCase.CaseType != models.CaseTypeCriminal {
		return nil
	}

	start, end, err := GetSlotTimeRange(slot, date)
	if err != nil {
		return []models.ConflictInfo{{Type: "error", Message: "invalid time slot"}}
	}

	type witnessInfo struct {
		WitnessType string
		WitnessName string
	}

	currentWitnesses := make(map[string]witnessInfo)
	for _, w := range currentCase.Witnesses {
		currentWitnesses[w.ID] = witnessInfo{
			WitnessType: w.WitnessType,
			WitnessName: w.Name,
		}
	}

	if len(currentWitnesses) == 0 {
		return nil
	}

	allHearings := store.GetAllHearings()
	for _, h := range allHearings {
		if h.ID == excludeHearingID {
			continue
		}
		if h.Status == models.HearingStatusCancelled || h.Status == models.HearingStatusPostponed {
			continue
		}
		if h.CaseID == caseID {
			continue
		}

		otherCase, ok := store.GetCase(h.CaseID)
		if !ok {
			continue
		}
		if otherCase.CaseType != models.CaseTypeCriminal {
			continue
		}

		hStart, hEnd, _ := GetSlotTimeRange(h.TimeSlot, h.Date)
		if !timesOverlap(start, end, hStart, hEnd) {
			continue
		}

		for _, otherW := range otherCase.Witnesses {
			if _, exists := currentWitnesses[otherW.ID]; exists {
				conflicts = append(conflicts, models.ConflictInfo{
					Type:    "witness",
					Message: fmt.Sprintf("证人 %s 同时出现在另一起刑事案件中，存在串供风险", otherW.Name),
					Detail:  fmt.Sprintf("与庭期 %s 中证人 %s (类型: %s) 时间重叠", h.ID, otherW.Name, otherW.WitnessType),
				})
			}
		}
	}

	return conflicts
}

func timesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

func CheckCourtroomOccupied(courtroomID string, date string, slot models.TimeSlot, excludeHearingID string) bool {
	hearings := store.GetAllHearings()
	for _, h := range hearings {
		if h.ID == excludeHearingID {
			continue
		}
		if h.CourtroomID != courtroomID {
			continue
		}
		if h.Date != date {
			continue
		}
		if h.TimeSlot != slot {
			continue
		}
		if h.Status == models.HearingStatusCancelled || h.Status == models.HearingStatusPostponed {
			continue
		}
		return true
	}
	return false
}

func GetJudgeDailyHearingCount(judgeID string, date string, excludeHearingID string) int {
	count := 0
	hearings := store.GetAllHearings()
	for _, h := range hearings {
		if h.ID == excludeHearingID {
			continue
		}
		if h.JudgeID != judgeID {
			continue
		}
		if h.Date != date {
			continue
		}
		if h.Status == models.HearingStatusCancelled || h.Status == models.HearingStatusPostponed {
			continue
		}
		count++
	}
	return count
}

func GetMaxHearingsPerDay(judgeID string, caseType models.CaseType) int {
	if caseType == models.CaseTypeCriminal {
		return 3
	}
	return 4
}
