package status

import (
	"fmt"
	"sort"
	"time"

	"example.com/apiary-colony-service/internal/domain"
	"example.com/apiary-colony-service/internal/store"
)

// DefaultInspectionWindow 是调用方未指定时使用的待巡检判定窗口。
const DefaultInspectionWindow = 14 * 24 * time.Hour

type HiveSummary struct {
	HiveID           string     `json:"hiveId"`
	State            string     `json:"state"`
	LastInspectionAt *time.Time `json:"lastInspectionAt"`
	LatestConcerns   []string   `json:"latestConcerns"`
	QueenID          string     `json:"queenId"`
	QueenMarkColor   string     `json:"queenMarkColor"`
}

type Report struct {
	ApiaryID         string         `json:"apiaryId"`
	Name             string         `json:"name"`
	AsOf             time.Time      `json:"asOf"`
	InspectionWindow string         `json:"inspectionWindow"`
	HiveCount        int            `json:"hiveCount"`
	StateCounts      map[string]int `json:"stateCounts"`
	NeedsInspection  []string       `json:"needsInspection"`
	ConcernHives     []string       `json:"concernHives"`
	LayingQueens     int            `json:"layingQueens"`
	Hives            []HiveSummary  `json:"hives"`
}

// BuildReport 在单份快照上按同一个 asOf 与巡检窗口完成聚合，
// 保证计数、待巡检集合与逐蜂箱摘要来自同一视图。
func BuildReport(snapshot store.Snapshot, apiary domain.Apiary, asOf time.Time, window time.Duration) Report {
	asOf = asOf.UTC()
	report := Report{
		ApiaryID:         apiary.ID,
		Name:             apiary.Name,
		AsOf:             asOf,
		InspectionWindow: FormatWindow(window),
		StateCounts:      initialStateCounts(),
		NeedsInspection:  []string{},
		ConcernHives:     []string{},
		Hives:            []HiveSummary{},
	}
	hives := snapshot.ApiaryHives(apiary.ID)
	report.HiveCount = len(hives)
	for _, hive := range hives {
		report.StateCounts[string(hive.State)]++
		summary := HiveSummary{
			HiveID:           hive.ID,
			State:            string(hive.State),
			LastInspectionAt: hive.LastInspectionAt,
			LatestConcerns:   []string{},
		}
		if queen, ok := snapshot.LayingQueen(hive.ID); ok {
			summary.QueenID = queen.ID
			summary.QueenMarkColor = queen.MarkColor
			report.LayingQueens++
		}
		if latest, ok := snapshot.LatestInspection(hive.ID); ok {
			observedAt := latest.ObservedAt
			summary.LastInspectionAt = &observedAt
			if len(latest.Concerns) > 0 {
				summary.LatestConcerns = append([]string{}, latest.Concerns...)
				report.ConcernHives = append(report.ConcernHives, hive.ID)
			}
		}
		if NeedsInspection(hive, asOf, window) {
			report.NeedsInspection = append(report.NeedsInspection, hive.ID)
		}
		report.Hives = append(report.Hives, summary)
	}
	sort.Strings(report.NeedsInspection)
	sort.Strings(report.ConcernHives)
	return report
}

// NeedsInspection 判定蜂箱是否需要巡检：从未巡检，或距 asOf 超过给定巡检窗口。
func NeedsInspection(hive domain.Hive, asOf time.Time, window time.Duration) bool {
	if hive.LastInspectionAt == nil {
		return true
	}
	return asOf.Sub(hive.LastInspectionAt.UTC()) > window
}

// FormatWindow 以可复现的紧凑形式回显巡检窗口：整天数写作 "14d"，
// 整小时写作 "36h"，其余退化为标准时长字符串；回显值可原样作为请求参数再次使用。
func FormatWindow(window time.Duration) string {
	if window > 0 {
		if window%(24*time.Hour) == 0 {
			return fmt.Sprintf("%dd", window/(24*time.Hour))
		}
		if window%time.Hour == 0 {
			return fmt.Sprintf("%dh", window/time.Hour)
		}
	}
	return window.String()
}

func initialStateCounts() map[string]int {
	return map[string]int{
		string(domain.HiveColonized):  0,
		string(domain.HiveProduction): 0,
		string(domain.HiveWintering):  0,
		string(domain.HiveRetired):    0,
	}
}
