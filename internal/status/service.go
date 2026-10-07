package status

import (
	"time"

	"example.com/apiary-colony-service/internal/domain"
	"example.com/apiary-colony-service/internal/store"
)

// Query 是单次状态汇总读取的参数：asOf 与巡检窗口共同决定待巡检判定，
// 两者缺省时分别取服务当前时间与 DefaultInspectionWindow。
type Query struct {
	AsOf   *time.Time
	Window *time.Duration
}

// Service 读取单份快照并给出蜂场状态汇总。
type Service struct {
	store *store.Memory
	clock func() time.Time
}

func New(memory *store.Memory, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{store: memory, clock: clock}
}

func (s *Service) ApiaryStatus(apiaryID string, query Query) (Report, error) {
	snapshot := s.store.Snapshot()
	apiary, ok := snapshot.Apiaries[apiaryID]
	if !ok {
		return Report{}, domain.NotFound("蜂场", apiaryID)
	}
	moment := s.clock()
	if query.AsOf != nil {
		moment = *query.AsOf
	}
	window := DefaultInspectionWindow
	if query.Window != nil {
		if *query.Window <= 0 {
			return Report{}, domain.Invalid("inspectionWindowDays", "巡检窗口必须为正数天数")
		}
		window = *query.Window
	}
	return BuildReport(snapshot, apiary, moment, window), nil
}
