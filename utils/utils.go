package utils

import (
	"fmt"
	"time"
)

func ToFloat(v any) float64 {
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

func ToFloatPtr(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	f := ToFloat(v)
	return &f
}

func ToTime(v any) time.Time {
	t, _ := v.(time.Time)
	return t
}

func PtrString(s interface{}) *string {
	if s == nil {
		return nil
	}
	str := fmt.Sprintf("%v", s)
	return &str
}
