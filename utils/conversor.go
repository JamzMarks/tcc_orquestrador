package utils

import "time"

func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func toTime(v any) time.Time {
	t, _ := v.(time.Time)
	return t
}
