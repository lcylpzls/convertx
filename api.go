package convertx

import (
	"time"

	"github.com/lcylpzls/convertx/internal/core"
)

// Version 是当前库版本，与 git tag 保持一致。
const Version = core.Version

// 错误码常量（errx.Code 类型）。
const (
	// ErrConversionFailed 通用转换失败：类型不兼容、格式非法等。
	ErrConversionFailed = core.ErrConversionFailed
	// ErrValidationFailed 校验失败：validx 校验规则不通过。
	ErrValidationFailed = core.ErrValidationFailed
	// ErrOverflow 数值溢出：源值超出目标类型范围。
	ErrOverflow = core.ErrOverflow
	// ErrUnknownFormat 时间格式无法识别。
	ErrUnknownFormat = core.ErrUnknownFormat
	// ErrCircularReference 结构体转换中检测到循环引用。
	ErrCircularReference = core.ErrCircularReference
	// ErrDuplicateConverter 同一 (源类型, 目标类型) 转换器重复注册。
	ErrDuplicateConverter = core.ErrDuplicateConverter
	// ErrNilPointer 开启 nil 指针报错模式时遇到 nil 指针。
	ErrNilPointer = core.ErrNilPointer
	// ErrUnmatchedField 严格模式下结构体/map 存在未匹配字段。
	ErrUnmatchedField = core.ErrUnmatchedField
)

// 时间戳精度常量。
const (
	// UnitSecond 秒。
	UnitSecond = core.UnitSecond
	// UnitMillisecond 毫秒。
	UnitMillisecond = core.UnitMillisecond
	// UnitMicrosecond 微秒。
	UnitMicrosecond = core.UnitMicrosecond
	// UnitNanosecond 纳秒。
	UnitNanosecond = core.UnitNanosecond
)

type (
	// Option 是转换选项函数类型。
	Option = core.Option
	// TimeUnit 是时间戳精度枚举。
	TimeUnit = core.TimeUnit
)

// To 将任意源值转换为目标类型 T。
func To[T any](src any, opts ...Option) (T, error) { return core.To[T](src, opts...) }

// ToOrDefault 转换失败或源为零值/nil 时返回默认值，永不返回错误。
func ToOrDefault[T any](src any, def T) T { return core.ToOrDefault(src, def) }

// 类型特定转换函数：为 To[T] 的便捷别名，行为完全一致。

func ToInt(src any, opts ...Option) (int, error)       { return core.ToInt(src, opts...) }
func ToInt8(src any, opts ...Option) (int8, error)     { return core.ToInt8(src, opts...) }
func ToInt16(src any, opts ...Option) (int16, error)   { return core.ToInt16(src, opts...) }
func ToInt32(src any, opts ...Option) (int32, error)   { return core.ToInt32(src, opts...) }
func ToInt64(src any, opts ...Option) (int64, error)   { return core.ToInt64(src, opts...) }
func ToUint(src any, opts ...Option) (uint, error)     { return core.ToUint(src, opts...) }
func ToUint8(src any, opts ...Option) (uint8, error)   { return core.ToUint8(src, opts...) }
func ToUint16(src any, opts ...Option) (uint16, error) { return core.ToUint16(src, opts...) }
func ToUint32(src any, opts ...Option) (uint32, error) { return core.ToUint32(src, opts...) }
func ToUint64(src any, opts ...Option) (uint64, error) { return core.ToUint64(src, opts...) }
func ToFloat32(src any, opts ...Option) (float32, error) {
	return core.ToFloat32(src, opts...)
}
func ToFloat64(src any, opts ...Option) (float64, error) {
	return core.ToFloat64(src, opts...)
}
func ToBool(src any, opts ...Option) (bool, error)     { return core.ToBool(src, opts...) }
func ToString(src any, opts ...Option) (string, error) { return core.ToString(src, opts...) }
func ToByte(src any, opts ...Option) (byte, error)     { return core.ToByte(src, opts...) }
func ToRune(src any, opts ...Option) (rune, error)     { return core.ToRune(src, opts...) }

// Struct 将源结构体（或 map）的字段复制/填充到目标结构体。
func Struct(src any, dst any, opts ...Option) error { return core.Struct(src, dst, opts...) }

// ToStruct 泛型版本的结构体转换。
func ToStruct[T any](src any, opts ...Option) (T, error) {
	return core.ToStruct[T](src, opts...)
}

// StructToMap 将源结构体导出为 map[string]any。
func StructToMap(src any, opts ...Option) (map[string]any, error) {
	return core.StructToMap(src, opts...)
}

// MapToStruct 将 map 填充到目标结构体，nil map 视为空 map。
func MapToStruct(src any, dst any, opts ...Option) error {
	return core.MapToStruct(src, dst, opts...)
}

// SliceToSlice 将源 slice 逐元素转换为目标 slice。
func SliceToSlice[Src any, Dst any](src []Src, opts ...Option) ([]Dst, error) {
	return core.SliceToSlice[Src, Dst](src, opts...)
}

// MapToMap 将源 map 的 key 与 value 分别转换。
func MapToMap[SK comparable, SV any, DK comparable, DV any](src map[SK]SV, opts ...Option) (map[DK]DV, error) {
	return core.MapToMap[SK, SV, DK, DV](src, opts...)
}

// SliceToMap 将源 slice 按 keyFn 提取的 key 转为 map。
func SliceToMap[T any, K comparable](src []T, keyFn func(T) K, opts ...Option) (map[K]T, error) {
	return core.SliceToMap(src, keyFn, opts...)
}

// MapToSlice 将源 map 的值提取为 slice，顺序不保证。
func MapToSlice[K comparable, V any](src map[K]V, opts ...Option) ([]V, error) {
	return core.MapToSlice[K, V](src, opts...)
}

// ToTime 将 string / 整数时间戳 / time.Time 转为 time.Time。
func ToTime(src any, opts ...Option) (time.Time, error) {
	return core.ToTime(src, opts...)
}

// TimeToString 按指定布局格式化时间。
func TimeToString(t time.Time, layout string) string { return core.TimeToString(t, layout) }

// TimestampToTime 将指定精度的时间戳转为 UTC 时间。
func TimestampToTime(ts int64, unit TimeUnit) time.Time {
	return core.TimestampToTime(ts, unit)
}

// TimeToTimestamp 将时间转为指定精度的时间戳。
func TimeToTimestamp(t time.Time, unit TimeUnit) int64 {
	return core.TimeToTimestamp(t, unit)
}

// ToUTC 转换为 UTC 时区。
func ToUTC(t time.Time) time.Time { return core.ToUTC(t) }

// ToLocal 转换为本地时区。
func ToLocal(t time.Time) time.Time { return core.ToLocal(t) }

// ToLocation 转换为指定时区。
func ToLocation(t time.Time, loc *time.Location) time.Time { return core.ToLocation(t, loc) }

// ToDuration 将 string / 整数（纳秒）/ time.Duration 转为 time.Duration。
func ToDuration(src any, opts ...Option) (time.Duration, error) {
	return core.ToDuration(src, opts...)
}

// ToAndValidate 转换成功后自动对结果执行 validx 校验。
func ToAndValidate[T any](src any, rules string, opts ...Option) (T, error) {
	return core.ToAndValidate[T](src, rules, opts...)
}

// Validate 对已有值执行 validx 校验。
func Validate(value any, rules string) error { return core.Validate(value, rules) }

// ValidateField 单字段校验，行为同 Validate。
func ValidateField(value any, rules string) error { return core.ValidateField(value, rules) }

// RegisterConverter 注册自定义类型转换器，优先级高于内置转换器。
func RegisterConverter[Src any, Dst any](fn func(Src) (Dst, error)) error {
	return core.RegisterConverter[Src, Dst](fn)
}
