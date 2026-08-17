package core

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// 数值 Kind 判断辅助。

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func isUintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

func isFloatKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func isNumericKind(k reflect.Kind) bool {
	return isIntKind(k) || isUintKind(k) || isFloatKind(k)
}

// convertScalar 将已解引用的标量源值转换为目标标量类型。
func convertScalar(srcVal reflect.Value, srcType, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	srcKind := srcType.Kind()
	dstKind := dstType.Kind()
	srcValI := srcVal.Interface()

	// 字符串作为源。
	if srcKind == reflect.String {
		s := strings.TrimSpace(srcVal.String())
		switch {
		case isNumericKind(dstKind):
			return parseNumericString(s, srcType, dstType, fieldPath)
		case dstKind == reflect.Bool:
			b, err := parseBoolString(s, srcType, dstType, fieldPath)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(b).Convert(dstType), nil
		}
	}

	// 字符串作为目标。
	if dstKind == reflect.String {
		switch {
		case isNumericKind(srcKind):
			return reflect.ValueOf(fmt.Sprintf("%v", srcValI)).Convert(dstType), nil
		case srcKind == reflect.Bool:
			if srcVal.Bool() {
				return reflect.ValueOf("true").Convert(dstType), nil
			}
			return reflect.ValueOf("false").Convert(dstType), nil
		}
	}

	// bool 与数值互转。
	if srcKind == reflect.Bool && isNumericKind(dstKind) {
		if srcVal.Bool() {
			return reflect.ValueOf(int64(1)).Convert(dstType), nil
		}
		return reflect.ValueOf(int64(0)).Convert(dstType), nil
	}
	if isNumericKind(srcKind) && dstKind == reflect.Bool {
		if !srcVal.IsZero() {
			return reflect.ValueOf(true).Convert(dstType), nil
		}
		return reflect.ValueOf(false).Convert(dstType), nil
	}

	// 数值与数值互转。
	if isNumericKind(srcKind) && isNumericKind(dstKind) {
		return convertNumber(srcVal, srcType, dstType, fieldPath)
	}

	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的转换：%s → %s", srcType, dstType),
		srcType, dstType, fieldPath, srcValI)
}

// parseNumericString 将字符串解析为目标数值类型，含溢出检测。
func parseNumericString(s string, srcType, dstType reflect.Type, fieldPath string) (reflect.Value, error) {
	switch dstType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, newError(ErrConversionFailed,
				fmt.Sprintf("无法将字符串 %q 解析为整数", s), srcType, dstType, fieldPath, s)
		}
		if !inIntRange(v, dstType.Kind()) {
			return reflect.Value{}, overflowError(v, srcType, dstType, fieldPath, s)
		}
		return reflect.ValueOf(v).Convert(dstType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// 负数字符串转无符号按溢出处理。
		if strings.HasPrefix(strings.TrimSpace(s), "-") {
			if v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
				return reflect.Value{}, overflowError(v, srcType, dstType, fieldPath, s)
			}
		}
		v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return reflect.Value{}, newError(ErrConversionFailed,
				fmt.Sprintf("无法将字符串 %q 解析为无符号整数", s), srcType, dstType, fieldPath, s)
		}
		if !inUintRange(v, dstType.Kind()) {
			return reflect.Value{}, overflowError(v, srcType, dstType, fieldPath, s)
		}
		return reflect.ValueOf(v).Convert(dstType), nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, newError(ErrConversionFailed,
				fmt.Sprintf("无法将字符串 %q 解析为浮点数", s), srcType, dstType, fieldPath, s)
		}
		if dstType.Kind() == reflect.Float32 && math.IsInf(float64(float32(v)), 0) && !math.IsInf(v, 0) {
			return reflect.Value{}, overflowError(v, srcType, dstType, fieldPath, s)
		}
		return reflect.ValueOf(v).Convert(dstType), nil
	}
	// 兜底：调用方已保证目标为数值类型，此路径仅防御直接调用。
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("目标类型 %s 不是数值类型", dstType), srcType, dstType, fieldPath, s)
}

// parseBoolString 解析布尔字符串，仅识别 true/false/1/0/t/f（大小写不敏感）。
func parseBoolString(s string, srcType, dstType reflect.Type, fieldPath string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "t":
		return true, nil
	case "false", "0", "f":
		return false, nil
	}
	return false, newError(ErrConversionFailed,
		fmt.Sprintf("无法将字符串 %q 解析为布尔值", s), srcType, dstType, fieldPath, s)
}

// convertNumber 执行数值类型之间的转换，含溢出检测。
func convertNumber(srcVal reflect.Value, srcType, dstType reflect.Type, fieldPath string) (reflect.Value, error) {
	srcKind := srcVal.Kind()
	dstKind := dstType.Kind()
	srcValI := srcVal.Interface()

	// 浮点 → 整数。
	if isFloatKind(srcKind) && isIntKind(dstKind) {
		f := srcVal.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return reflect.Value{}, newError(ErrConversionFailed,
				"NaN/Inf 不能转换为整数", srcType, dstType, fieldPath, srcValI)
		}
		if f < -math.Exp2(63) || f >= math.Exp2(63) || !inIntRange(int64(f), dstKind) {
			return reflect.Value{}, overflowError(f, srcType, dstType, fieldPath, srcValI)
		}
		return reflect.ValueOf(int64(f)).Convert(dstType), nil
	}
	// 浮点 → 无符号。
	if isFloatKind(srcKind) && isUintKind(dstKind) {
		f := srcVal.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f >= math.Exp2(64) || !inUintRange(uint64(f), dstKind) {
			return reflect.Value{}, overflowError(f, srcType, dstType, fieldPath, srcValI)
		}
		return reflect.ValueOf(uint64(f)).Convert(dstType), nil
	}
	// 有符号 → 无符号。
	if isIntKind(srcKind) && isUintKind(dstKind) {
		i := srcVal.Int()
		if i < 0 || !inUintRange(uint64(i), dstKind) {
			return reflect.Value{}, overflowError(i, srcType, dstType, fieldPath, srcValI)
		}
		return reflect.ValueOf(i).Convert(dstType), nil
	}
	// 无符号 → 有符号。
	if isUintKind(srcKind) && isIntKind(dstKind) {
		u := srcVal.Uint()
		if u > math.MaxInt64 || !inIntRange(int64(u), dstKind) {
			return reflect.Value{}, overflowError(u, srcType, dstType, fieldPath, srcValI)
		}
		return reflect.ValueOf(u).Convert(dstType), nil
	}
	// 无符号 → 无符号。
	if isUintKind(srcKind) && isUintKind(dstKind) {
		u := srcVal.Uint()
		if !inUintRange(u, dstKind) {
			return reflect.Value{}, overflowError(u, srcType, dstType, fieldPath, srcValI)
		}
		return reflect.ValueOf(u).Convert(dstType), nil
	}
	// 有符号 → 有符号。
	if isIntKind(srcKind) && isIntKind(dstKind) {
		i := srcVal.Int()
		if !inIntRange(i, dstKind) {
			return reflect.Value{}, overflowError(i, srcType, dstType, fieldPath, srcValI)
		}
		return reflect.ValueOf(i).Convert(dstType), nil
	}
	// 整数 → 浮点：直接转换（允许精度损失）。
	if (isIntKind(srcKind) || isUintKind(srcKind)) && isFloatKind(dstKind) {
		return srcVal.Convert(dstType), nil
	}
	// 浮点 → 浮点。
	if isFloatKind(srcKind) && isFloatKind(dstKind) {
		f := srcVal.Float()
		if dstKind == reflect.Float32 {
			f32 := float32(f)
			if math.IsInf(float64(f32), 0) && !math.IsInf(f, 0) {
				return reflect.Value{}, overflowError(f, srcType, dstType, fieldPath, srcValI)
			}
			return reflect.ValueOf(f32), nil
		}
		return srcVal.Convert(dstType), nil
	}
	// 兜底：调用方已保证源与目标均为数值类型，此路径仅防御直接调用。
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的数值转换：%s → %s", srcType, dstType), srcType, dstType, fieldPath, srcValI)
}

// inIntRange 判断 int64 是否在目标整数类型的范围内。
func inIntRange(v int64, kind reflect.Kind) bool {
	switch kind {
	case reflect.Int8:
		return v >= math.MinInt8 && v <= math.MaxInt8
	case reflect.Int16:
		return v >= math.MinInt16 && v <= math.MaxInt16
	case reflect.Int32:
		return v >= math.MinInt32 && v <= math.MaxInt32
	default:
		return true // Int / Int64：int64 本身范围。
	}
}

// inUintRange 判断 uint64 是否在目标无符号类型的范围内。
func inUintRange(v uint64, kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint8:
		return v <= math.MaxUint8
	case reflect.Uint16:
		return v <= math.MaxUint16
	case reflect.Uint32:
		return v <= math.MaxUint32
	case reflect.Uintptr:
		return v <= uint64(^uintptr(0))
	default:
		return true // Uint / Uint64：uint64 本身范围。
	}
}

// overflowError 构造数值溢出错误。
func overflowError(v any, srcType, dstType reflect.Type, fieldPath string, src any) error {
	return newError(ErrOverflow,
		fmt.Sprintf("数值溢出：%v 超出 %s 范围", v, dstType), srcType, dstType, fieldPath, src)
}
