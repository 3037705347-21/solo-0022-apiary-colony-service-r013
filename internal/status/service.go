package status

import (
	"time"

	"example.com/apiary-colony-service/internal/domain"
	"example.com/apiary-colony-service/internal/store"
)

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

// ApiaryStatus 在单份快照上按同一个时间点与巡检窗口聚合蜂场状态。
// asOf 缺省取服务当前时间；window 缺省取 DefaultInspectionWindow。
// 显式传入的非正窗口返回校验错误，绝不静默回退到默认窗口。
func (s *Service) ApiaryStatus(apiaryID string, asOf *time.Time, window *time.Duration) (Report, error) {
	snapshot := s.store.Snapshot()
	apiary, ok := snapshot.Apiaries[apiaryID]
	if !ok {
		return Report{}, domain.NotFound("蜂场", apiaryID)
	}
	moment := s.clock()
	if asOf != nil {
		moment = *asOf
	}
	inspectionWindow := DefaultInspectionWindow
	if window != nil {
		if *window <= 0 {
			return Report{}, domain.Invalid("inspectionWindow", "巡检窗口必须为正时长")
		}
		inspectionWindow = *window
	}
	return BuildReport(snapshot, apiary, moment, inspectionWindow), nil
}
