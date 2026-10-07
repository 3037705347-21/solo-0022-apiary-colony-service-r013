package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"example.com/apiary-colony-service/internal/domain"
)

const maxBodyBytes = 1 << 20

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return domain.Invalid("body", "请求体不能为空")
		}
		return domain.Invalid("body", "请求体不是合法 JSON：%v", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.Invalid("body", "请求体只能包含一个 JSON 对象")
	}
	return nil
}

// parseOptionalTime 解析可缺省的时间参数，缺省时返回 nil 交由调用方取当前时间。
func parseOptionalTime(field, value string) (*time.Time, error) {
	trimmed := domain.TrimText(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, domain.Invalid(field, "%s 必须是 RFC3339 时间", field)
	}
	utc := parsed.UTC()
	return &utc, nil
}

// maxInspectionWindowDays 限制巡检窗口上限，避免换算成 Duration 时溢出。
const maxInspectionWindowDays = 3650

// parseOptionalInspectionWindow 解析可缺省的巡检窗口天数，缺省时返回 nil
// 交由调用方取缺省窗口；非法取值返回校验错误而不是静默回退。
func parseOptionalInspectionWindow(field, value string) (*time.Duration, error) {
	trimmed := domain.TrimText(value)
	if trimmed == "" {
		return nil, nil
	}
	days, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil, domain.Invalid(field, "%s 必须是正整数天数", field)
	}
	if days < 1 || days > maxInspectionWindowDays {
		return nil, domain.Invalid(field, "%s 必须在 1 到 %d 天之间", field, maxInspectionWindowDays)
	}
	window := time.Duration(days) * 24 * time.Hour
	return &window, nil
}
