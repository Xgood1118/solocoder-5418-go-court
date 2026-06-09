package services

import (
	"fmt"
	"time"

	"court-scheduler/models"
	"court-scheduler/store"
)

func GenerateMonthlyStats(month string) models.MonthlyStats {
	courtroomUtil := calculateCourtroomUtilization(month)
	judgeWorkloads := calculateJudgeWorkloads(month)
	avgDuration := calculateAvgDuration(month)
	postponeRate := calculatePostponeRate(month)

	stats := models.MonthlyStats{
		Month:          month,
		CourtroomUtil:  courtroomUtil,
		JudgeWorkloads: judgeWorkloads,
		AvgDuration:    avgDuration,
		PostponeRate:   postponeRate,
		GeneratedAt:    time.Now(),
	}

	store.SaveStats(stats)
	return stats
}

func calculateCourtroomUtilization(month string) map[string]models.CourtroomUtilization {
	result := make(map[string]models.CourtroomUtilization)

	year, monthInt := parseMonth(month)
	daysInMonth := daysInMonth(year, monthInt)
	totalSlots := daysInMonth * 4

	courtrooms := store.GetAllCourtrooms()
	for _, cr := range courtrooms {
		result[cr.ID] = models.CourtroomUtilization{
			TotalSlots:  totalSlots,
			UsedSlots:   0,
			Utilization: 0,
		}
	}

	hearings := store.GetAllHearings()
	for _, h := range hearings {
		hYear, hMonth, _ := parseDateToYMD(h.Date)
		if fmt.Sprintf("%d-%02d", hYear, hMonth) != month {
			continue
		}
		if h.Status == models.HearingStatusCancelled {
			continue
		}

		util, ok := result[h.CourtroomID]
		if ok {
			util.UsedSlots++
			if util.TotalSlots > 0 {
				util.Utilization = float64(util.UsedSlots) / float64(util.TotalSlots)
			}
			result[h.CourtroomID] = util
		}
	}

	return result
}

func calculateJudgeWorkloads(month string) []models.JudgeWorkload {
	workloadMap := make(map[string]*models.JudgeWorkload)

	judges := store.GetAllJudges()
	for _, j := range judges {
		workloadMap[j.ID] = &models.JudgeWorkload{
			JudgeID:    j.ID,
			JudgeName:  j.Name,
			TotalCases: 0,
			ByType:     make(map[models.CaseType]int),
		}
	}

	hearings := store.GetAllHearings()
	for _, h := range hearings {
		hYear, hMonth, _ := parseDateToYMD(h.Date)
		if fmt.Sprintf("%d-%02d", hYear, hMonth) != month {
			continue
		}
		if h.Status == models.HearingStatusCancelled {
			continue
		}

		wl, ok := workloadMap[h.JudgeID]
		if ok {
			wl.TotalCases++
			caseData, cok := store.GetCase(h.CaseID)
			if cok {
				wl.ByType[caseData.CaseType]++
			}
		}
	}

	result := []models.JudgeWorkload{}
	for _, wl := range workloadMap {
		result = append(result, *wl)
	}

	return result
}

func calculateAvgDuration(month string) models.AvgDurationReport {
	durationByType := make(map[models.CaseType]int)
	countByType := make(map[models.CaseType]int)

	hearings := store.GetAllHearings()
	for _, h := range hearings {
		hYear, hMonth, _ := parseDateToYMD(h.Date)
		if fmt.Sprintf("%d-%02d", hYear, hMonth) != month {
			continue
		}
		if h.Status == models.HearingStatusCancelled {
			continue
		}

		caseData, cok := store.GetCase(h.CaseID)
		if !cok {
			continue
		}

		duration := h.DurationMin
		if duration <= 0 {
			duration = int(h.EndTime.Sub(h.StartTime).Minutes())
		}

		durationByType[caseData.CaseType] += duration
		countByType[caseData.CaseType]++
	}

	avgByType := make(map[models.CaseType]float64)
	for ct, total := range durationByType {
		if countByType[ct] > 0 {
			avgByType[ct] = float64(total) / float64(countByType[ct])
		}
	}

	return models.AvgDurationReport{
		ByType: avgByType,
	}
}

func calculatePostponeRate(month string) models.PostponeRateReport {
	total := 0
	postponed := 0

	hearings := store.GetAllHearings()
	for _, h := range hearings {
		hYear, hMonth, _ := parseDateToYMD(h.Date)
		if fmt.Sprintf("%d-%02d", hYear, hMonth) != month {
			continue
		}

		total++
		if h.Status == models.HearingStatusPostponed {
			postponed++
		}
	}

	rate := 0.0
	if total > 0 {
		rate = float64(postponed) / float64(total)
	}

	return models.PostponeRateReport{
		TotalScheduled: total,
		Postponed:      postponed,
		Rate:           rate,
	}
}

func parseMonth(month string) (int, int) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		now := time.Now()
		return now.Year(), int(now.Month())
	}
	return t.Year(), int(t.Month())
}

func parseDateToYMD(dateStr string) (int, int, int) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		now := time.Now()
		return now.Year(), int(now.Month()), now.Day()
	}
	return t.Year(), int(t.Month()), t.Day()
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()
}

func StartMonthlyStatsScheduler() {
	go func() {
		for {
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local)
			waitDuration := nextRun.Sub(now)

			fmt.Printf("next monthly stats generation in %v\n", waitDuration)
			time.Sleep(waitDuration)

			prevMonth := now.AddDate(0, -1, 0)
			prevMonthStr := fmt.Sprintf("%d-%02d", prevMonth.Year(), int(prevMonth.Month()))
			GenerateMonthlyStats(prevMonthStr)
			fmt.Printf("generated monthly stats for %s\n", prevMonthStr)
		}
	}()
	fmt.Println("monthly stats scheduler started")
}
