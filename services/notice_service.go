package services

import (
	"fmt"
	"time"

	"court-scheduler/models"
	"court-scheduler/store"

	"github.com/google/uuid"
)

var NoticeTypeConfig = map[models.NoticeType]map[string]interface{}{
	models.NoticeTypeJudgeOA: {
		"name":        "法官内部OA",
		"description": "通过法院内部OA系统发送给法官",
		"channel":     "oa",
	},
	models.NoticeTypeLawyerSMS: {
		"name":        "律师短信",
		"description": "通过短信发送给律师",
		"channel":     "sms",
	},
	models.NoticeTypePartyEMS: {
		"name":        "当事人EMS",
		"description": "通过EMS邮政快递寄送纸质通知给当事人",
		"channel":     "ems",
	},
	models.NoticeTypeWitnessCall: {
		"name":        "证人电话",
		"description": "通过电话通知证人",
		"channel":     "phone",
	},
	models.NoticeTypeExpertEmail: {
		"name":        "鉴定人邮件",
		"description": "通过邮件通知鉴定人",
		"channel":     "email",
	},
	models.NoticeTypeJurorSMS: {
		"name":        "陪审员短信",
		"description": "通过短信通知人民陪审员",
		"channel":     "sms",
	},
	models.NoticeTypeTranslatorEmail: {
		"name":        "翻译邮件",
		"description": "通过邮件通知翻译人员",
		"channel":     "email",
	},
}

func DispatchNotices(hearingID string) {
	hearing, ok := store.GetHearing(hearingID)
	if !ok {
		fmt.Printf("hearing %s not found, cannot dispatch notices\n", hearingID)
		return
	}

	caseData, ok := store.GetCase(hearing.CaseID)
	if !ok {
		fmt.Printf("case %s not found, cannot dispatch notices\n", hearing.CaseID)
		return
	}

	judge, _ := store.GetJudge(hearing.JudgeID)
	sendNotice(hearingID, hearing.JudgeID, judge.Name, models.NoticeTypeJudgeOA,
		fmt.Sprintf("您有新的庭期安排：%s，时间：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))

	for _, lawyer := range caseData.Lawyers {
		sendNotice(hearingID, lawyer.ID, lawyer.Name, models.NoticeTypeLawyerSMS,
			fmt.Sprintf("【法院通知】您代理的案件%s有新庭期，时间：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	for _, party := range caseData.Parties {
		sendNotice(hearingID, party.ID, party.Name, models.NoticeTypePartyEMS,
			fmt.Sprintf("【法院传票】案件%s，开庭时间：%s %s，请准时到庭。", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	for _, witness := range caseData.Witnesses {
		sendNotice(hearingID, witness.ID, witness.Name, models.NoticeTypeWitnessCall,
			fmt.Sprintf("【法院通知】案件%s，请您作为证人出庭，时间：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	if hearing.Expert != "" {
		sendNotice(hearingID, "expert_"+hearing.Expert, hearing.Expert, models.NoticeTypeExpertEmail,
			fmt.Sprintf("【法院通知】案件%s鉴定人出庭通知，时间：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	if hearing.Translator != "" {
		sendNotice(hearingID, "translator_"+hearing.Translator, hearing.Translator, models.NoticeTypeTranslatorEmail,
			fmt.Sprintf("【法院通知】案件%s翻译人员出庭通知，时间：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	for _, jurorID := range hearing.JurorIDs {
		juror, _ := store.GetJuror(jurorID)
		sendNotice(hearingID, jurorID, juror.Name, models.NoticeTypeJurorSMS,
			fmt.Sprintf("【人民陪审员通知】您被抽选参与案件%s陪审，时间：%s %s，请准时到庭。", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	fmt.Printf("dispatched notices for hearing %s\n", hearingID)
}

func DispatchPostponeNotices(hearingID string) {
	hearing, ok := store.GetHearing(hearingID)
	if !ok {
		return
	}

	caseData, ok := store.GetCase(hearing.CaseID)
	if !ok {
		return
	}

	allNotices := store.GetAllNotices()
	for _, n := range allNotices {
		if n.HearingID == hearingID && !n.IsReminder {
			n.FollowUp = true
			store.SaveNotice(n)
		}
	}

	judge, _ := store.GetJudge(hearing.JudgeID)
	sendNotice(hearingID, hearing.JudgeID, judge.Name, models.NoticeTypeJudgeOA,
		fmt.Sprintf("【延期通知】案件%s庭期已变更，新时间：%s %s，原因：%s",
			caseData.Title, hearing.Date, hearing.TimeSlot, hearing.PostponeReason))

	for _, lawyer := range caseData.Lawyers {
		sendNotice(hearingID, lawyer.ID, lawyer.Name, models.NoticeTypeLawyerSMS,
			fmt.Sprintf("【法院延期通知】案件%s庭期变更，新时间：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	for _, party := range caseData.Parties {
		sendNotice(hearingID, party.ID, party.Name, models.NoticeTypePartyEMS,
			fmt.Sprintf("【法院延期传票】案件%s开庭时间变更为：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	for _, witness := range caseData.Witnesses {
		sendNotice(hearingID, witness.ID, witness.Name, models.NoticeTypeWitnessCall,
			fmt.Sprintf("【法院延期通知】案件%s证人出庭时间变更为：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	if hearing.Expert != "" {
		sendNotice(hearingID, "expert_"+hearing.Expert, hearing.Expert, models.NoticeTypeExpertEmail,
			fmt.Sprintf("【法院延期通知】案件%s鉴定人出庭时间变更：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	if hearing.Translator != "" {
		sendNotice(hearingID, "translator_"+hearing.Translator, hearing.Translator, models.NoticeTypeTranslatorEmail,
			fmt.Sprintf("【法院延期通知】案件%s翻译出庭时间变更：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}

	for _, jurorID := range hearing.JurorIDs {
		juror, _ := store.GetJuror(jurorID)
		sendNotice(hearingID, jurorID, juror.Name, models.NoticeTypeJurorSMS,
			fmt.Sprintf("【陪审员延期通知】案件%s陪审时间变更为：%s %s", caseData.Title, hearing.Date, hearing.TimeSlot))
	}
}

func sendNotice(hearingID, recipientID, recipientName string, noticeType models.NoticeType, content string) {
	now := time.Now()
	notice := models.Notice{
		ID:            uuid.New().String(),
		HearingID:     hearingID,
		RecipientID:   recipientID,
		RecipientName: recipientName,
		NoticeType:    noticeType,
		Status:        models.NoticeStatusSent,
		Content:       content,
		SentAt:        now,
		FollowUp:      false,
		IsReminder:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	store.SaveNotice(notice)
}

func SendFollowUpReminders() {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	tomorrowStr := tomorrow.Format("2006-01-02")

	hearings := store.GetAllHearings()
	for _, h := range hearings {
		if h.Date != tomorrowStr {
			continue
		}
		if h.Status == models.HearingStatusCancelled || h.Status == models.HearingStatusPostponed {
			continue
		}

		notices := store.GetAllNotices()
		for _, n := range notices {
			if n.HearingID != h.ID {
				continue
			}
			if n.IsReminder {
				continue
			}
			if n.Status == models.NoticeStatusConfirmed {
				continue
			}

			caseData, _ := store.GetCase(h.CaseID)
			reminderContent := fmt.Sprintf("【开庭提醒】明天%s %s，案件%s，请准时到庭。", h.Date, h.TimeSlot, caseData.Title)

			reminder := models.Notice{
				ID:            uuid.New().String(),
				HearingID:     h.ID,
				RecipientID:   n.RecipientID,
				RecipientName: n.RecipientName,
				NoticeType:    n.NoticeType,
				Status:        models.NoticeStatusSent,
				Content:       reminderContent,
				SentAt:        time.Now(),
				FollowUp:      false,
				IsReminder:    true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			store.SaveNotice(reminder)

			n.FollowUp = true
			store.SaveNotice(n)
		}
	}
}

func StartFollowUpScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			SendFollowUpReminders()
		}
	}()
	fmt.Println("follow-up reminder scheduler started")
}
