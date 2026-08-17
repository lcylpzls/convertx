package core

import (
	"testing"

	"github.com/lcylpzls/testx"
)

func TestToOrDefault(t *testing.T) {
	// 转换成功返回转换值。
	v := ToOrDefault[int]("42", -1)
	testx.Equal(t, v, 42)

	// 转换失败返回默认值。
	v = ToOrDefault[int]("abc", -1)
	testx.Equal(t, v, -1)

	// 源为零值（"0"）触发默认值（设计决策）。
	v = ToOrDefault[int]("0", -1)
	testx.Equal(t, v, -1)

	// 源为 nil 返回默认值。
	s := ToOrDefault[string](nil, "N/A")
	testx.Equal(t, s, "N/A")

	// 字符串默认值成功路径。
	s = ToOrDefault[string](123, "N/A")
	testx.Equal(t, s, "123")

	// 非零成功路径不触发默认值。
	f := ToOrDefault[float64]("1.5", 0)
	testx.Equal(t, f, 1.5)
}

func TestToOrDefaultPointerZero(t *testing.T) {
	// nil 指针源 → 转换结果零值 → 触发默认值。
	var p *int
	v := ToOrDefault[int](p, -1)
	testx.Equal(t, v, -1)
}

func TestWithDefaultOption(t *testing.T) {
	// 转换失败时返回默认值，error 为 nil。
	v, err := To[int]("abc", WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)

	// 转换成功返回转换值。
	v, err = To[int]("42", WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, 42)

	// 源为 nil 返回默认值。
	v, err = To[int](nil, WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)

	// 源为 nil 指针返回默认值。
	var p *int
	v, err = To[int](p, WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)
}

func TestIsZeroValue(t *testing.T) {
	testx.True(t, isZeroValue(0))
	testx.False(t, isZeroValue(1))
	testx.True(t, isZeroValue(""))
	testx.False(t, isZeroValue("x"))
	testx.True(t, isZeroValue(false))
	testx.True(t, isZeroValue(nil))
}

func TestWithDefaultZeroValue(t *testing.T) {
	// 成功转换结果为零值时，同样按 WithDefault 语义返回默认值。
	v, err := To[int]("0", WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)
}

func TestWithDefaultConvertibleValue(t *testing.T) {
	// 默认值会经过类型转换，而不是原样返回。
	v, err := To[int]("abc", WithDefault("42"))
	testx.NoError(t, err)
	testx.Equal(t, v, 42)
}

func TestWithDefaultWrongType(t *testing.T) {
	// 默认值无法转换为目标类型时返回明确错误，而不是类型断言失败。
	_, err := To[int]("abc", WithDefault("fallback"))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestToAnyDefault(t *testing.T) {
	v, err := To[any](nil, WithDefault(1))
	testx.NoError(t, err)
	testx.Equal(t, v, 1)

	// 目标为 any 时，零值源同样触发默认值。
	v, err = To[any]("", WithDefault("d"))
	testx.NoError(t, err)
	testx.Equal(t, v, "d")
}

func TestStructWithDefault(t *testing.T) {
	type srcT struct {
		Age int
	}
	type dstT struct {
		Age int
	}
	src := srcT{}
	var dst dstT
	err := Struct(&src, &dst, WithDefault(dstT{Age: 9}))
	testx.NoError(t, err)
	testx.Equal(t, dst.Age, 9)
}

func TestStructWithValidation(t *testing.T) {
	type srcT struct {
		Age int
	}
	type dstT struct {
		Age int
	}
	src := srcT{Age: 150}
	var dst dstT
	err := Struct(&src, &dst, WithValidation("min=1"))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrValidationFailed)
}
