package core

import (
	"errors"
	"math"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

func TestToInt(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want int
	}{
		{"字符串", "42", 42},
		{"int64", int64(7), 7},
		{"float 截断", 3.9, 3},
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"uint", uint(9), 9},
		{"byte", byte(5), 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[int](tt.src)
			testx.NoError(t, err)
			testx.Equal(t, got, tt.want)
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want string
	}{
		{"整数", 42, "42"},
		{"浮点", 1.5, "1.5"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"字符串原样", "abc", "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[string](tt.src)
			testx.NoError(t, err)
			testx.Equal(t, got, tt.want)
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want bool
	}{
		{"true 字符串", "true", true},
		{"大写 TRUE", "TRUE", true},
		{"t", "t", true},
		{"1", "1", true},
		{"false", "false", false},
		{"f", "F", false},
		{"0", "0", false},
		{"数值非零", 42, true},
		{"数值零", 0, false},
		{"float 非零", 1.5, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := To[bool](tt.src)
			testx.NoError(t, err)
			testx.Equal(t, got, tt.want)
		})
	}
}

func TestToFloat(t *testing.T) {
	got, err := To[float64]("3.14")
	testx.NoError(t, err)
	testx.Equal(t, got, 3.14)

	got32, err := To[float32](2.5)
	testx.NoError(t, err)
	testx.Equal(t, got32, float32(2.5))
}

func TestToByteAndRune(t *testing.T) {
	b, err := To[byte]("65")
	testx.NoError(t, err)
	testx.Equal(t, b, byte(65))

	r, err := To[rune]("97")
	testx.NoError(t, err)
	testx.Equal(t, r, rune(97))
}

func TestConvertErrors(t *testing.T) {
	// 非法字符串转整数。
	_, err := To[int]("abc")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	// 溢出：128 超出 int8 范围。
	_, err = To[int8](128)
	testx.ErrCode(t, err, ErrOverflow)

	// 溢出：-1 转无符号。
	_, err = To[uint](int(-1))
	testx.ErrCode(t, err, ErrOverflow)

	// 非法字符串转 bool。
	_, err = To[bool]("maybe")
	testx.ErrCode(t, err, ErrConversionFailed)

	// 不支持的类型组合（struct → int）。
	_, err = To[int](struct{}{})
	testx.ErrCode(t, err, ErrConversionFailed)

	// 字符串 "1.5" 转 int。
	_, err = To[int]("1.5")
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestErrorContext(t *testing.T) {
	_, err := To[int8]("abc")
	var e *errx.Error
	testx.True(t, errors.As(err, &e))
	testx.ErrFields(t, err,
		errx.KV{Key: "source_type", Value: "string"},
		errx.KV{Key: "target_type", Value: "int8"},
		errx.KV{Key: "field_path", Value: ""},
	)
}

func TestErrorValuePreviewTruncated(t *testing.T) {
	long := string(make([]byte, 200))
	_, err := To[int](long)
	var e *errx.Error
	testx.True(t, errors.As(err, &e))
	for _, kv := range e.Fields() {
		if kv.Key == "value_preview" {
			s, ok := kv.Value.(string)
			testx.True(t, ok)
			testx.Equal(t, len(s), valuePreviewLimit)
		}
	}
}

func TestIntRangeBoundaries(t *testing.T) {
	// int8 边界。
	v, err := To[int8](127)
	testx.NoError(t, err)
	testx.Equal(t, v, int8(127))

	_, err = To[int8](128)
	testx.ErrCode(t, err, ErrOverflow)

	v, err = To[int8](-128)
	testx.NoError(t, err)
	testx.Equal(t, v, int8(-128))

	// uint8 边界。
	uv, err := To[uint8](255)
	testx.NoError(t, err)
	testx.Equal(t, uv, uint8(255))

	_, err = To[uint8](256)
	testx.ErrCode(t, err, ErrOverflow)

	// uint64 → int64 溢出。
	_, err = To[int64](uint64(math.MaxUint64))
	testx.ErrCode(t, err, ErrOverflow)
}

func TestFloatToIntBoundaries(t *testing.T) {
	_, err := To[int64](math.Exp2(63))
	testx.ErrCode(t, err, ErrOverflow)

	_, err = To[int64](math.NaN())
	testx.ErrCode(t, err, ErrConversionFailed)

	_, err = To[uint](math.Inf(1))
	testx.ErrCode(t, err, ErrOverflow)

	// float64 → float32 溢出。
	_, err = To[float32](math.Exp2(200))
	testx.ErrCode(t, err, ErrOverflow)
}

func TestUintToFloat(t *testing.T) {
	v, err := To[float64](uint64(42))
	testx.NoError(t, err)
	testx.Equal(t, v, float64(42))
}

func TestNegativeStringToUint(t *testing.T) {
	_, err := To[uint]("-5")
	testx.ErrCode(t, err, ErrOverflow)
}

func TestStringTrim(t *testing.T) {
	v, err := To[int]("  42  ")
	testx.NoError(t, err)
	testx.Equal(t, v, 42)
}
