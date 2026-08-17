package core

import "reflect"

// ToOrDefault 将 src 转换为目标类型 T；转换失败或源为零值/nil 时返回默认值 def。
// 永不返回错误，调用方无需处理 error。
//
// 注意：源为类型零值（""、0、false 等）时同样触发默认值，
// 如需区分"零值"与"缺失"，请使用指针类型或标准转换 API 自行判断。
func ToOrDefault[T any](src any, def T) T {
	v, err := To[T](src)
	if err != nil {
		return def
	}
	if isZeroValue(v) {
		return def
	}
	return v
}

// isZeroValue 判断值是否为其类型的零值。
func isZeroValue(v any) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	return rv.IsZero()
}
