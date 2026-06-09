package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"court-scheduler/models"
)

type Store struct {
	Courtrooms *sync.Map
	Judges       *sync.Map
	Jurors       *sync.Map
	Cases        *sync.Map
	Hearings     *sync.Map
	Notices      *sync.Map
	Stats        *sync.Map

	ScheduleMutex sync.Mutex

	SnapshotDir string
}

var GlobalStore *Store

func NewStore(snapshotDir string) *Store {
	s := &Store{
		Courtrooms:  &sync.Map{},
		Judges:      &sync.Map{},
		Jurors:      &sync.Map{},
		Cases:       &sync.Map{},
		Hearings:    &sync.Map{},
		Notices:     &sync.Map{},
		Stats:       &sync.Map{},
		SnapshotDir: snapshotDir,
	}
	return s
}

func InitStore(snapshotDir string) {
	GlobalStore = NewStore(snapshotDir)
	if err := GlobalStore.LoadLatestSnapshot(); err != nil {
		fmt.Printf("no snapshot found or load failed: %v\n", err)
	}
}

func (s *Store) LoadLatestSnapshot() error {
	files, err := os.ReadDir(s.SnapshotDir)
	if err != nil {
		return err
	}

	var snapshotFiles []os.DirEntry
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			snapshotFiles = append(snapshotFiles, f)
		}
	}

	if len(snapshotFiles) == 0 {
		return fmt.Errorf("no snapshot files")
	}

	sort.Slice(snapshotFiles, func(i, j int) bool {
		return snapshotFiles[i].Name() > snapshotFiles[j].Name()
	})

	latestPath := filepath.Join(s.SnapshotDir, snapshotFiles[0].Name())
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return err
	}

	var snapshot models.SnapshotData
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	for k, v := range snapshot.Courtrooms {
		s.Courtrooms.Store(k, v)
	}
	for k, v := range snapshot.Judges {
		s.Judges.Store(k, v)
	}
	for k, v := range snapshot.Jurors {
		s.Jurors.Store(k, v)
	}
	for k, v := range snapshot.Cases {
		s.Cases.Store(k, v)
	}
	for k, v := range snapshot.Hearings {
		s.Hearings.Store(k, v)
	}
	for k, v := range snapshot.Notices {
		s.Notices.Store(k, v)
	}
	for k, v := range snapshot.Stats {
		s.Stats.Store(k, v)
	}

	fmt.Printf("loaded snapshot from %s\n", latestPath)
	return nil
}

func (s *Store) DumpSnapshot() error {
	snapshot := models.SnapshotData{
		Timestamp:  time.Now(),
		Courtrooms: make(map[string]models.Courtroom),
		Judges:     make(map[string]models.Judge),
		Jurors:     make(map[string]models.Juror),
		Cases:      make(map[string]models.Case),
		Hearings:   make(map[string]models.Hearing),
		Notices:    make(map[string]models.Notice),
		Stats:      make(map[string]models.MonthlyStats),
	}

	s.Courtrooms.Range(func(key, value interface{}) bool {
		snapshot.Courtrooms[key.(string)] = value.(models.Courtroom)
		return true
	})
	s.Judges.Range(func(key, value interface{}) bool {
		snapshot.Judges[key.(string)] = value.(models.Judge)
		return true
	})
	s.Jurors.Range(func(key, value interface{}) bool {
		snapshot.Jurors[key.(string)] = value.(models.Juror)
		return true
	})
	s.Cases.Range(func(key, value interface{}) bool {
		snapshot.Cases[key.(string)] = value.(models.Case)
		return true
	})
	s.Hearings.Range(func(key, value interface{}) bool {
		snapshot.Hearings[key.(string)] = value.(models.Hearing)
		return true
	})
	s.Notices.Range(func(key, value interface{}) bool {
		snapshot.Notices[key.(string)] = value.(models.Notice)
		return true
	})
	s.Stats.Range(func(key, value interface{}) bool {
		snapshot.Stats[key.(string)] = value.(models.MonthlyStats)
		return true
	})

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("snapshot_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(s.SnapshotDir, filename)

	if err := os.MkdirAll(s.SnapshotDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return err
	}

	fmt.Printf("snapshot saved to %s\n", filepath)
	return nil
}

func (s *Store) StartAutoSnapshot(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := s.DumpSnapshot(); err != nil {
				fmt.Printf("snapshot dump failed: %v\n", err)
			}
		}
	}()
	fmt.Println("auto snapshot started, interval:", interval)
}

func syncMapToMap(m *sync.Map) map[string]interface{} {
	result := make(map[string]interface{})
	m.Range(func(key, value interface{}) bool {
		result[key.(string)] = value
		return true
	})
	return result
}

func GetAllCourtrooms() []models.Courtroom {
	var result []models.Courtroom
	GlobalStore.Courtrooms.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.Courtroom))
		return true
	})
	return result
}

func GetCourtroom(id string) (models.Courtroom, bool) {
	v, ok := GlobalStore.Courtrooms.Load(id)
	if !ok {
		return models.Courtroom{}, false
	}
	return v.(models.Courtroom), true
}

func SaveCourtroom(c models.Courtroom) {
	GlobalStore.Courtrooms.Store(c.ID, c)
}

func DeleteCourtroom(id string) {
	GlobalStore.Courtrooms.Delete(id)
}

func GetAllJudges() []models.Judge {
	var result []models.Judge
	GlobalStore.Judges.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.Judge))
		return true
	})
	return result
}

func GetJudge(id string) (models.Judge, bool) {
	v, ok := GlobalStore.Judges.Load(id)
	if !ok {
		return models.Judge{}, false
	}
	return v.(models.Judge), true
}

func SaveJudge(j models.Judge) {
	GlobalStore.Judges.Store(j.ID, j)
}

func DeleteJudge(id string) {
	GlobalStore.Judges.Delete(id)
}

func GetAllJurors() []models.Juror {
	var result []models.Juror
	GlobalStore.Jurors.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.Juror))
		return true
	})
	return result
}

func GetJuror(id string) (models.Juror, bool) {
	v, ok := GlobalStore.Jurors.Load(id)
	if !ok {
		return models.Juror{}, false
	}
	return v.(models.Juror), true
}

func SaveJuror(j models.Juror) {
	GlobalStore.Jurors.Store(j.ID, j)
}

func DeleteJuror(id string) {
	GlobalStore.Jurors.Delete(id)
}

func GetAllCases() []models.Case {
	var result []models.Case
	GlobalStore.Cases.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.Case))
		return true
	})
	return result
}

func GetCase(id string) (models.Case, bool) {
	v, ok := GlobalStore.Cases.Load(id)
	if !ok {
		return models.Case{}, false
	}
	return v.(models.Case), true
}

func SaveCase(c models.Case) {
	GlobalStore.Cases.Store(c.ID, c)
}

func DeleteCase(id string) {
	GlobalStore.Cases.Delete(id)
}

func GetAllHearings() []models.Hearing {
	var result []models.Hearing
	GlobalStore.Hearings.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.Hearing))
		return true
	})
	return result
}

func GetHearing(id string) (models.Hearing, bool) {
	v, ok := GlobalStore.Hearings.Load(id)
	if !ok {
		return models.Hearing{}, false
	}
	return v.(models.Hearing), true
}

func SaveHearing(h models.Hearing) {
	GlobalStore.Hearings.Store(h.ID, h)
}

func DeleteHearing(id string) {
	GlobalStore.Hearings.Delete(id)
}

func GetAllNotices() []models.Notice {
	var result []models.Notice
	GlobalStore.Notices.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.Notice))
		return true
	})
	return result
}

func GetNotice(id string) (models.Notice, bool) {
	v, ok := GlobalStore.Notices.Load(id)
	if !ok {
		return models.Notice{}, false
	}
	return v.(models.Notice), true
}

func SaveNotice(n models.Notice) {
	GlobalStore.Notices.Store(n.ID, n)
}

func DeleteNotice(id string) {
	GlobalStore.Notices.Delete(id)
}

func GetAllStats() []models.MonthlyStats {
	var result []models.MonthlyStats
	GlobalStore.Stats.Range(func(key, value interface{}) bool {
		result = append(result, value.(models.MonthlyStats))
		return true
	})
	return result
}

func GetStats(month string) (models.MonthlyStats, bool) {
	v, ok := GlobalStore.Stats.Load(month)
	if !ok {
		return models.MonthlyStats{}, false
	}
	return v.(models.MonthlyStats), true
}

func SaveStats(s models.MonthlyStats) {
	GlobalStore.Stats.Store(s.Month, s)
}
