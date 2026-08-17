package core

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	// timeType 是 time.Time 的反射类型。
	timeType = reflect.TypeOf(time.Time{})
	// durationType 是 time.Duration 的反射类型。
	durationType = reflect.TypeOf(time.Duration(0))
)

// timeLayouts 是 string → time.Time 自动识别的格式列表，按优先级排列。
var timeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02",
	time.RFC822,
	time.RFC822Z,
}

// timezoneSuffixRe 匹配字符串末尾的时区偏移（如 +08:00、-0700、Z）。
var timezoneSuffixRe = regexp.MustCompile(`[+-]\d{2}:?\d{2}$`)

// convertToTimeValue 将源值转换为 time.Time。
func convertToTimeValue(srcVal reflect.Value, srcType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	switch srcType.Kind() {
	case reflect.String:
		t, err := parseTimeString(strings.TrimSpace(srcVal.String()), cfg)
		if err != nil {
			return reflect.Value{}, withFieldPath(err, fieldPath)
		}
		return reflect.ValueOf(t), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		ts, err := toInt64(srcVal)
		if err != nil {
			return reflect.Value{}, withFieldPath(err, fieldPath)
		}
		unit := cfg.timeUnit
		if !cfg.timeUnitSet {
			unit = inferTimeUnit(ts)
		}
		return reflect.ValueOf(timestampToTime(ts, unit)), nil
	}
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的时间源类型：%s", srcType), srcType, timeType, fieldPath, srcVal.Interface())
}

// convertToDurationValue 将源值转换为 time.Duration。
func convertToDurationValue(srcVal reflect.Value, srcType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	switch srcType.Kind() {
	case reflect.String:
		d, err := time.ParseDuration(strings.TrimSpace(srcVal.String()))
		if err != nil {
			return reflect.Value{}, newError(ErrConversionFailed,
				fmt.Sprintf("无法解析持续时间字符串 %q", srcVal.String()),
				srcType, durationType, fieldPath, srcVal.String())
		}
		return reflect.ValueOf(d), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		nsec, err := toInt64(srcVal)
		if err != nil {
			return reflect.Value{}, withFieldPath(err, fieldPath)
		}
		return reflect.ValueOf(time.Duration(nsec)), nil
	}
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的持续时间源类型：%s", srcType), srcType, durationType, fieldPath, srcVal.Interface())
}

// convertFromTime 将 time.Time / time.Duration 源值转换为其他类型。
func convertFromTime(srcVal reflect.Value, srcType, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	if srcType == timeType {
		return convertFromTimeValue(srcVal, dstType, cfg, fieldPath)
	}
	return convertFromDurationValue(srcVal, dstType, cfg, fieldPath)
}

// convertFromTimeValue 将 time.Time 转为 string（默认 RFC3339）或整数（时间戳）。
func convertFromTimeValue(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	t := srcVal.Interface().(time.Time)
	switch dstType.Kind() {
	case reflect.String:
		layout := cfg.timeLayout
		if layout == "" {
			layout = time.RFC3339
		}
		return reflect.ValueOf(t.Format(layout)), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		unit := cfg.timeUnit
		if !cfg.timeUnitSet {
			unit = UnitSecond
		}
		ts := timeToTimestamp(t, unit)
		if !int64InRange(ts, dstType.Kind()) {
			return reflect.Value{}, overflowError(ts, timeType, dstType, fieldPath, t)
		}
		return reflect.ValueOf(ts).Convert(dstType), nil
	}
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的时间目标类型：%s", dstType), timeType, dstType, fieldPath, t)
}

// convertFromDurationValue 将 time.Duration 转为 string 或整数（纳秒）。
func convertFromDurationValue(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	d := srcVal.Interface().(time.Duration)
	switch dstType.Kind() {
	case reflect.String:
		return reflect.ValueOf(d.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if !int64InRange(int64(d), dstType.Kind()) {
			return reflect.Value{}, overflowError(int64(d), durationType, dstType, fieldPath, d)
		}
		return reflect.ValueOf(int64(d)).Convert(dstType), nil
	}
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的持续时间目标类型：%s", dstType), durationType, dstType, fieldPath, d)
}

// parseTimeString 解析时间字符串：指定格式时仅尝试该格式，否则按优先级自动识别。
func parseTimeString(s string, cfg *convertConfig) (time.Time, error) {
	// 指定格式：仅尝试该格式，不回退。
	if cfg.timeLayout != "" {
		t, err := time.Parse(cfg.timeLayout, s)
		if err != nil {
			return time.Time{}, newError(ErrUnknownFormat,
				fmt.Sprintf("无法按布局 %q 解析时间 %q", cfg.timeLayout, s), nil, timeType, "", s)
		}
		return applyDefaultLocation(t, s, cfg), nil
	}
	// 自动识别。
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return applyDefaultLocation(t, s, cfg), nil
		}
	}
	return time.Time{}, newError(ErrUnknownFormat,
		fmt.Sprintf("无法识别时间格式：%q", s), nil, timeType, "", s)
}

// applyDefaultLocation 对无时区偏移的解析结果应用默认时区（默认 time.Local）。
// 带时区偏移的字符串保留原时区。
func applyDefaultLocation(t time.Time, s string, cfg *convertConfig) time.Time {
	if hasTimezoneOffset(s) {
		return t
	}
	loc := cfg.defaultLocation
	if !cfg.defaultLocationSet || loc == nil {
		loc = time.Local
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

// hasTimezoneOffset 判断时间字符串是否携带时区偏移标记。
func hasTimezoneOffset(s string) bool {
	if strings.HasSuffix(s, "Z") || strings.HasSuffix(s, "z") {
		return true
	}
	return timezoneSuffixRe.MatchString(s)
}

// inferTimeUnit 按数值范围自动推断时间戳精度。
func inferTimeUnit(ts int64) TimeUnit {
	abs := ts
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs < 1e11:
		return UnitSecond
	case abs < 1e14:
		return UnitMillisecond
	case abs < 1e17:
		return UnitMicrosecond
	default:
		return UnitNanosecond
	}
}

// timestampToTime 将指定精度的时间戳转为 UTC 时间。
func timestampToTime(ts int64, unit TimeUnit) time.Time {
	switch unit {
	case UnitSecond:
		return time.Unix(ts, 0).UTC()
	case UnitMillisecond:
		return time.UnixMilli(ts).UTC()
	case UnitMicrosecond:
		return time.UnixMicro(ts).UTC()
	default:
		return time.Unix(0, ts).UTC()
	}
}

// timeToTimestamp 将时间转为指定精度的时间戳。
func timeToTimestamp(t time.Time, unit TimeUnit) int64 {
	switch unit {
	case UnitSecond:
		return t.Unix()
	case UnitMillisecond:
		return t.UnixMilli()
	case UnitMicrosecond:
		return t.UnixMicro()
	default:
		return t.UnixNano()
	}
}

// toInt64 将整数类型的反射值转为 int64，无符号大值溢出时报错。
func toInt64(srcVal reflect.Value) (int64, error) {
	switch srcVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return srcVal.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := srcVal.Uint()
		if u > math.MaxInt64 {
			return 0, overflowError(u, srcVal.Type(), timeType, "", srcVal.Interface())
		}
		return int64(u), nil
	}
	return 0, newError(ErrConversionFailed,
		fmt.Sprintf("无法将 %s 转为整数", srcVal.Type()), srcVal.Type(), nil, "", srcVal.Interface())
}

// int64InRange 判断 int64 是否在目标整数类型范围内（含无符号目标）。
func int64InRange(v int64, kind reflect.Kind) bool {
	if isUintKind(kind) {
		return v >= 0 && inUintRange(uint64(v), kind)
	}
	return inIntRange(v, kind)
}
