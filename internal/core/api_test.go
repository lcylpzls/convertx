package core

import (
	"testing"

	"github.com/lcylpzls/testx"
)

// TestTypedConversionFunctions 覆盖全部类型特定转换函数的成功与失败路径。
func TestTypedConversionFunctions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (any, error)
		want any
	}{
		{"ToInt", func() (any, error) { return ToInt("42") }, 42},
		{"ToInt8", func() (any, error) { return ToInt8("42") }, int8(42)},
		{"ToInt16", func() (any, error) { return ToInt16("42") }, int16(42)},
		{"ToInt32", func() (any, error) { return ToInt32("42") }, int32(42)},
		{"ToInt64", func() (any, error) { return ToInt64("42") }, int64(42)},
		{"ToUint", func() (any, error) { return ToUint("42") }, uint(42)},
		{"ToUint8", func() (any, error) { return ToUint8("42") }, uint8(42)},
		{"ToUint16", func() (any, error) { return ToUint16("42") }, uint16(42)},
		{"ToUint32", func() (any, error) { return ToUint32("42") }, uint32(42)},
		{"ToUint64", func() (any, error) { return ToUint64("42") }, uint64(42)},
		{"ToFloat32", func() (any, error) { return ToFloat32("1.5") }, float32(1.5)},
		{"ToFloat64", func() (any, error) { return ToFloat64("1.5") }, 1.5},
		{"ToBool", func() (any, error) { return ToBool("true") }, true},
		{"ToString", func() (any, error) { return ToString(42) }, "42"},
		{"ToByte", func() (any, error) { return ToByte("65") }, byte(65)},
		{"ToRune", func() (any, error) { return ToRune("97") }, rune(97)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn()
			testx.NoError(t, err)
			testx.Equal(t, got, tt.want)
		})
	}
}

func TestTypedConversionFunctionsErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"ToInt", func() error { _, err := ToInt("abc"); return err }},
		{"ToInt8", func() error { _, err := ToInt8(128); return err }},
		{"ToInt16", func() error { _, err := ToInt16(40000); return err }},
		{"ToInt32", func() error { _, err := ToInt32("x"); return err }},
		{"ToInt64", func() error { _, err := ToInt64("x"); return err }},
		{"ToUint", func() error { _, err := ToUint(-1); return err }},
		{"ToUint8", func() error { _, err := ToUint8(256); return err }},
		{"ToUint16", func() error { _, err := ToUint16(70000); return err }},
		{"ToUint32", func() error { _, err := ToUint32("x"); return err }},
		{"ToUint64", func() error { _, err := ToUint64("x"); return err }},
		{"ToFloat32", func() error { _, err := ToFloat32("x"); return err }},
		{"ToFloat64", func() error { _, err := ToFloat64("x"); return err }},
		{"ToBool", func() error { _, err := ToBool("x"); return err }},
		{"ToString", func() error { _, err := ToString(struct{}{}); return err }},
		{"ToByte", func() error { _, err := ToByte(300); return err }},
		{"ToRune", func() error { _, err := ToRune(1 << 40); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testx.Error(t, tt.fn())
		})
	}
}

func TestWithTagNameOption(t *testing.T) {
	type src struct {
		Name string `json:"nickname"`
	}
	type dst struct {
		Nickname string `json:"nickname"`
	}
	srcVal := src{Name: "ivan"}
	var dstVal dst
	err := Struct(&srcVal, &dstVal, WithTagName("json"))
	testx.NoError(t, err)
	testx.Equal(t, dstVal.Nickname, "ivan")
}

func TestOptionsCombination(t *testing.T) {
	// 默认值 + 校验组合：校验失败时返回默认值。
	v, err := To[int]("200", WithValidation("min=0,max=100"), WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)
}

func TestNilOptionSkipped(t *testing.T) {
	v, err := To[int]("42", nil)
	testx.NoError(t, err)
	testx.Equal(t, v, 42)
}
