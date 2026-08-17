package core

import (
	"fmt"
	"reflect"
	"time"
)

// To 将任意源值转换为目标类型 T 的泛型转换入口。
// 转换失败返回 T 的零值与 errx 封装的错误；禁止 panic、禁止吞错误。
func To[T any](src any, opts ...Option) (T, error) {
	var zero T
	dstType := reflect.TypeOf(zero)
	v, err := Convert(src, dstType, opts...)
	if err != nil {
		return zero, err
	}
	if v == nil {
		return zero, nil
	}
	out, ok := v.(T)
	if !ok {
		return zero, newError(ErrConversionFailed, "转换结果类型断言失败", nil, dstType, "", src)
	}
	return out, nil
}

// 类型特定转换函数：为 To[T] 的便捷别名，行为完全一致。

func ToInt(src any, opts ...Option) (int, error)       { return To[int](src, opts...) }
func ToInt8(src any, opts ...Option) (int8, error)     { return To[int8](src, opts...) }
func ToInt16(src any, opts ...Option) (int16, error)   { return To[int16](src, opts...) }
func ToInt32(src any, opts ...Option) (int32, error)   { return To[int32](src, opts...) }
func ToInt64(src any, opts ...Option) (int64, error)   { return To[int64](src, opts...) }
func ToUint(src any, opts ...Option) (uint, error)     { return To[uint](src, opts...) }
func ToUint8(src any, opts ...Option) (uint8, error)   { return To[uint8](src, opts...) }
func ToUint16(src any, opts ...Option) (uint16, error) { return To[uint16](src, opts...) }
func ToUint32(src any, opts ...Option) (uint32, error) { return To[uint32](src, opts...) }
func ToUint64(src any, opts ...Option) (uint64, error) { return To[uint64](src, opts...) }
func ToFloat32(src any, opts ...Option) (float32, error) {
	return To[float32](src, opts...)
}
func ToFloat64(src any, opts ...Option) (float64, error) {
	return To[float64](src, opts...)
}
func ToBool(src any, opts ...Option) (bool, error)     { return To[bool](src, opts...) }
func ToString(src any, opts ...Option) (string, error) { return To[string](src, opts...) }
func ToByte(src any, opts ...Option) (byte, error)     { return To[byte](src, opts...) }
func ToRune(src any, opts ...Option) (rune, error)     { return To[rune](src, opts...) }

// Struct 将源结构体（或 map）的字段复制/填充到目标结构体。
// dst 必须为非 nil 结构体指针；源为 nil 指针时目标各字段保持零值。
func Struct(src any, dst any, opts ...Option) error {
	if dst == nil {
		return newError(ErrConversionFailed, "目标不能为 nil", nil, nil, "", src)
	}
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return newError(ErrConversionFailed, "目标必须是结构体指针", nil, nil, "", src)
	}
	if dv.Elem().Kind() != reflect.Struct {
		return newError(ErrConversionFailed, "目标必须是结构体指针", nil, nil, "", src)
	}

	cfg := defaultConfig()
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	if src == nil {
		if cfg.defaultSet {
			return setDefaultToStruct(dv.Elem(), &cfg)
		}
		return nil
	}
	sv := reflect.ValueOf(src)
	if hasNilPointer(sv) {
		if cfg.defaultSet {
			return setDefaultToStruct(dv.Elem(), &cfg)
		}
		if cfg.nilPointerAsError {
			return newError(ErrNilPointer, "源值为 nil 指针", sv.Type(), dv.Elem().Type(), "", src)
		}
		return nil
	}

	out, err := convert(sv, dv.Elem().Type(), &cfg, "")
	if err != nil {
		if cfg.defaultSet {
			return setDefaultToStruct(dv.Elem(), &cfg)
		}
		return err
	}
	if cfg.defaultSet && isZeroValue(out.Interface()) {
		return setDefaultToStruct(dv.Elem(), &cfg)
	}
	if cfg.validationSet {
		if verr := validateValue(out.Interface(), cfg.validationRules); verr != nil {
			if cfg.defaultSet {
				return setDefaultToStruct(dv.Elem(), &cfg)
			}
			return verr
		}
	}
	dv.Elem().Set(out)
	return nil
}

// setDefaultToStruct 将 WithDefault 的默认值转换后写入目标结构体。
func setDefaultToStruct(dst reflect.Value, cfg *convertConfig) error {
	dv, err := convertDefaultValue(cfg, dst.Type())
	if err != nil {
		return err
	}
	dst.Set(reflect.ValueOf(dv))
	return nil
}

// ToStruct 泛型版本的结构体转换，返回目标结构体值。
func ToStruct[T any](src any, opts ...Option) (T, error) {
	return To[T](src, opts...)
}

// StructToMap 将源结构体导出为 map[string]any。
func StructToMap(src any, opts ...Option) (map[string]any, error) {
	m, err := To[map[string]any](src, opts...)
	return m, err
}

// MapToStruct 将 map（支持强类型 map，如 map[string]string）填充到目标结构体。
// nil map 视为空 map。
func MapToStruct(src any, dst any, opts ...Option) error {
	return Struct(src, dst, opts...)
}

// SliceToSlice 将源 slice 逐元素转换为目标 slice，保留原顺序。
// 任一元素转换失败则整体失败；nil slice 返回 nil。
func SliceToSlice[Src any, Dst any](src []Src, opts ...Option) ([]Dst, error) {
	sliceType := reflect.SliceOf(reflect.TypeOf((*Dst)(nil)).Elem())
	out, err := Convert(src, sliceType, opts...)
	if err != nil {
		return nil, err
	}
	return out.([]Dst), nil
}

// MapToMap 将源 map 的 key 与 value 分别转换，返回目标 map。
// 任一 key/value 转换失败则整体失败；nil map 返回 nil。
func MapToMap[SK comparable, SV any, DK comparable, DV any](src map[SK]SV, opts ...Option) (map[DK]DV, error) {
	mapType := reflect.MapOf(reflect.TypeOf((*DK)(nil)).Elem(), reflect.TypeOf((*DV)(nil)).Elem())
	out, err := Convert(src, mapType, opts...)
	if err != nil {
		return nil, err
	}
	return out.(map[DK]DV), nil
}

// SliceToMap 将源 slice 按 keyFn 提取的 key 转为 map。
// key 重复时后者覆盖前者；keyFn panic 会被捕获并转为错误。
func SliceToMap[T any, K comparable](src []T, keyFn func(T) K, opts ...Option) (map[K]T, error) {
	m := make(map[K]T, len(src))
	for _, item := range src {
		k, err := callKeyFn(keyFn, item)
		if err != nil {
			return nil, err
		}
		m[k] = item
	}
	return m, nil
}

// callKeyFn 调用 keyFn 并防御 panic。
func callKeyFn[T any, K comparable](keyFn func(T) K, item T) (k K, err error) {
	defer func() {
		if r := recover(); r != nil {
			k = *new(K)
			err = newError(ErrConversionFailed, fmt.Sprintf("keyFn panic：%v", r), nil, nil, "", item)
		}
	}()
	return keyFn(item), nil
}

// MapToSlice 将源 map 的值提取为 slice，顺序不保证（Go map 迭代无序）。
func MapToSlice[K comparable, V any](src map[K]V, opts ...Option) ([]V, error) {
	out := make([]V, 0, len(src))
	for _, v := range src {
		out = append(out, v)
	}
	return out, nil
}

// ToTime 将 string / 整数时间戳 / time.Time 转为 time.Time。
func ToTime(src any, opts ...Option) (time.Time, error) {
	return To[time.Time](src, opts...)
}

// TimeToString 按指定布局格式化时间，永不失败。
func TimeToString(t time.Time, layout string) string {
	return t.Format(layout)
}

// TimestampToTime 将指定精度的时间戳转为 UTC 时间，永不失败。
func TimestampToTime(ts int64, unit TimeUnit) time.Time {
	return timestampToTime(ts, unit)
}

// TimeToTimestamp 将时间转为指定精度的时间戳。
func TimeToTimestamp(t time.Time, unit TimeUnit) int64 {
	return timeToTimestamp(t, unit)
}

// ToUTC 转换为 UTC 时区。
func ToUTC(t time.Time) time.Time { return t.UTC() }

// ToLocal 转换为本地时区。
func ToLocal(t time.Time) time.Time { return t.Local() }

// ToLocation 转换为指定时区。
func ToLocation(t time.Time, loc *time.Location) time.Time { return t.In(loc) }

// ToDuration 将 string（如 "1h30m"）/ 整数（纳秒）/ time.Duration 转为 time.Duration。
func ToDuration(src any, opts ...Option) (time.Duration, error) {
	return To[time.Duration](src, opts...)
}

// ToAndValidate 转换成功后自动对结果执行 validx 校验。
// 转换失败返回转换错误（不执行校验）；校验失败返回 ErrValidationFailed。
func ToAndValidate[T any](src any, rules string, opts ...Option) (T, error) {
	opts = append(opts, WithValidation(rules))
	return To[T](src, opts...)
}

// Validate 对已有值执行 validx 校验，不做转换。
// 校验通过返回 nil；失败返回 errx 封装的 ErrValidationFailed。
func Validate(value any, rules string) error {
	return validateValue(value, rules)
}

// ValidateField 单字段校验，行为同 Validate。
func ValidateField(value any, rules string) error {
	return validateValue(value, rules)
}
