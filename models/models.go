package models

import "time"

type CourtroomSize string

const (
	CourtroomSizeLarge  CourtroomSize = "large"
	CourtroomSizeMedium CourtroomSize = "medium"
	CourtroomSizeSmall  CourtroomSize = "small"
)

type CourtroomEquipment struct {
	Projector       bool `json:"projector"`
	Recording       bool `json:"recording"`
	Interpretation  bool `json:"interpretation"`
	DetentionAccess bool `json:"detention_access"`
}

type TimeSlot string

const (
	SlotMorning    TimeSlot = "morning"
	SlotAfternoon  TimeSlot = "afternoon"
	SlotEvening    TimeSlot = "evening"
	SlotWeekendAM  TimeSlot = "weekend_am"
)

type Courtroom struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Size        CourtroomSize       `json:"size"`
	Capacity    int                 `json:"capacity"`
	Equipment   CourtroomEquipment  `json:"equipment"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type CaseType string

const (
	CaseTypeCivil    CaseType = "civil"
	CaseTypeCriminal CaseType = "criminal"
	CaseTypeAdmin    CaseType = "administrative"
)

type Party struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"` // plaintiff, defendant, etc.
	Phone  string `json:"phone"`
	Email  string `json:"email"`
}

type Lawyer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Firm  string `json:"firm"`
}

type Witness struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	WitnessType string `json:"witness_type"` // plaintiff_witness, defendant_witness, expert
}

type Case struct {
	ID           string     `json:"id"`
	CaseNumber   string     `json:"case_number"`
	CaseType     CaseType   `json:"case_type"`
	Title        string     `json:"title"`
	Parties      []Party    `json:"parties"`
	Lawyers      []Lawyer   `json:"lawyers"`
	Witnesses    []Witness  `json:"witnesses"`
	CourtroomSize CourtroomSize `json:"courtroom_size"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Judge struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CaseTypes []CaseType `json:"case_types"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Juror struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CaseTypes []CaseType `json:"case_types"`
	Phone     string     `json:"phone"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type HearingStatus string

const (
	HearingStatusScheduled HearingStatus = "scheduled"
	HearingStatusPostponed HearingStatus = "postponed"
	HearingStatusCompleted HearingStatus = "completed"
	HearingStatusCancelled HearingStatus = "cancelled"
)

type JurorDrawAudit struct {
	DrawTime     time.Time `json:"draw_time"`
	CandidateIDs []string  `json:"candidate_ids"`
	SelectedIDs  []string  `json:"selected_ids"`
	Reason       string    `json:"reason"`
}

type Hearing struct {
	ID             string              `json:"id"`
	CaseID         string              `json:"case_id"`
	JudgeID        string              `json:"judge_id"`
	CourtroomID    string              `json:"courtroom_id"`
	Date           string              `json:"date"` // YYYY-MM-DD
	TimeSlot       TimeSlot            `json:"time_slot"`
	StartTime      time.Time           `json:"start_time"`
	EndTime        time.Time           `json:"end_time"`
	DurationMin    int                 `json:"duration_min"`
	Status         HearingStatus       `json:"status"`
	OriginalDate   string              `json:"original_date,omitempty"`
	OriginalSlot   TimeSlot            `json:"original_slot,omitempty"`
	PostponeReason string              `json:"postpone_reason,omitempty"`
	JurorIDs       []string            `json:"juror_ids"`
	JurorCount     int                 `json:"juror_count"`
	JurorDrawAudit JurorDrawAudit      `json:"juror_draw_audit"`
	Translator     string              `json:"translator,omitempty"`
	Expert         string              `json:"expert,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type NoticeType string

const (
	NoticeTypeJudgeOA   NoticeType = "judge_oa"
	NoticeTypeLawyerSMS NoticeType = "lawyer_sms"
	NoticeTypePartyEMS  NoticeType = "party_ems"
	NoticeTypeWitnessCall NoticeType = "witness_call"
	NoticeTypeExpertEmail NoticeType = "expert_email"
	NoticeTypeJurorSMS  NoticeType = "juror_sms"
	NoticeTypeTranslatorEmail NoticeType = "translator_email"
)

type NoticeStatus string

const (
	NoticeStatusSent     NoticeStatus = "sent"
	NoticeStatusRead     NoticeStatus = "read"
	NoticeStatusConfirmed NoticeStatus = "confirmed"
)

type Notice struct {
	ID            string       `json:"id"`
	HearingID     string       `json:"hearing_id"`
	RecipientID   string       `json:"recipient_id"`
	RecipientName string       `json:"recipient_name"`
	NoticeType    NoticeType   `json:"notice_type"`
	Status        NoticeStatus `json:"status"`
	Content       string       `json:"content"`
	SentAt        time.Time    `json:"sent_at"`
	ConfirmedAt   *time.Time   `json:"confirmed_at,omitempty"`
	FollowUp      bool         `json:"follow_up"`
	IsReminder    bool         `json:"is_reminder"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type ConflictInfo struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Detail   string `json:"detail,omitempty"`
}

type ScheduleWarning struct {
	Level   string `json:"level"` // yellow, red
	Message string `json:"message"`
}

type ScheduleResult struct {
	Success    bool              `json:"success"`
	Hearing    *Hearing          `json:"hearing,omitempty"`
	Conflicts  []ConflictInfo    `json:"conflicts,omitempty"`
	Warnings   []ScheduleWarning `json:"warnings,omitempty"`
}

type CourtroomUtilization struct {
	TotalSlots   int     `json:"total_slots"`
	UsedSlots    int     `json:"used_slots"`
	Utilization  float64 `json:"utilization"`
}

type JudgeWorkload struct {
	JudgeID     string   `json:"judge_id"`
	JudgeName   string   `json:"judge_name"`
	TotalCases  int      `json:"total_cases"`
	ByType      map[CaseType]int `json:"by_type"`
}

type AvgDurationReport struct {
	ByType map[CaseType]float64 `json:"by_type"`
}

type PostponeRateReport struct {
	TotalScheduled int     `json:"total_scheduled"`
	Postponed      int     `json:"postponed"`
	Rate           float64 `json:"rate"`
}

type MonthlyStats struct {
	Month              string                      `json:"month"`
	CourtroomUtil      map[string]CourtroomUtilization `json:"courtroom_utilization"`
	JudgeWorkloads     []JudgeWorkload             `json:"judge_workloads"`
	AvgDuration        AvgDurationReport           `json:"avg_duration"`
	PostponeRate       PostponeRateReport          `json:"postpone_rate"`
	GeneratedAt        time.Time                   `json:"generated_at"`
}

type SnapshotData struct {
	Timestamp    time.Time            `json:"timestamp"`
	Courtrooms   map[string]Courtroom `json:"courtrooms"`
	Judges       map[string]Judge     `json:"judges"`
	Jurors       map[string]Juror     `json:"jurors"`
	Cases        map[string]Case      `json:"cases"`
	Hearings     map[string]Hearing   `json:"hearings"`
	Notices      map[string]Notice    `json:"notices"`
	Stats        map[string]MonthlyStats `json:"stats"`
}
