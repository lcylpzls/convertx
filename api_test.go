// Package convertx_test 是根包黑盒冒烟测试，验证公开 API 层转发行为。
package convertx_test

import (
	"errors"
	"testing"
	"time"

	"github.com/lcylpzls/convertx"
	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

func TestVersionAndConstants(t *testing.T) {
	testx.Equal(t, convertx.Version, "v1.0.0")

	// 错误码常量。
	testx.Equal(t, convertx.ErrConversionFailed, errx.Code("CONVERTX_CONVERSION_FAILED"))
	testx.Equal(t, convertx.ErrValidationFailed, errx.Code("CONVERTX_VALIDATION_FAILED"))
	testx.Equal(t, convertx.ErrOverflow, errx.Code("CONVERTX_OVERFLOW"))
	testx.Equal(t, convertx.ErrUnknownFormat, errx.Code("CONVERTX_UNKNOWN_FORMAT"))
	testx.Equal(t, convertx.ErrCircularReference, errx.Code("CONVERTX_CIRCULAR_REFERENCE"))
	testx.Equal(t, convertx.ErrDuplicateConverter, errx.Code("CONVERTX_DUPLICATE_CONVERTER"))
	testx.Equal(t, convertx.ErrNilPointer, errx.Code("CONVERTX_NIL_POINTER"))
	testx.Equal(t, convertx.ErrUnmatchedField, errx.Code("CONVERTX_UNMATCHED_FIELD"))

	// 时间戳精度常量。
	testx.Equal(t, convertx.UnitSecond, convertx.TimeUnit(0))
	testx.Equal(t, convertx.UnitMillisecond, convertx.TimeUnit(1))
	testx.Equal(t, convertx.UnitMicrosecond, convertx.TimeUnit(2))
	testx.Equal(t, convertx.UnitNanosecond, convertx.TimeUnit(3))
}

func TestBasicConversion(t *testing.T) {
	// 泛型入口。
	v, err := convertx.To[int]("42")
	testx.NoError(t, err)
	testx.Equal(t, v, 42)

	// 类型特定函数。
	i, err := convertx.ToInt("7")
	testx.NoError(t, err)
	testx.Equal(t, i, 7)

	s, err := convertx.ToString(123)
	testx.NoError(t, err)
	testx.Equal(t, s, "123")

	b, err := convertx.ToBool("true")
	testx.NoError(t, err)
	testx.True(t, b)

	f, err := convertx.ToFloat64("1.5")
	testx.NoError(t, err)
	testx.Equal(t, f, 1.5)

	// 溢出错误。
	_, err = convertx.ToInt8(128)
	testx.Error(t, err)
	testx.ErrCode(t, err, convertx.ErrOverflow)
}

func TestDefaultValue(t *testing.T) {
	v := convertx.ToOrDefault[int]("abc", -1)
	testx.Equal(t, v, -1)

	v = convertx.ToOrDefault[int]("42", -1)
	testx.Equal(t, v, 42)
}

func TestAllTypedFunctions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (any, error)
		want any
	}{
		{"ToInt8", func() (any, error) { return convertx.ToInt8("8") }, int8(8)},
		{"ToInt16", func() (any, error) { return convertx.ToInt16("16") }, int16(16)},
		{"ToInt32", func() (any, error) { return convertx.ToInt32("32") }, int32(32)},
		{"ToInt64", func() (any, error) { return convertx.ToInt64("64") }, int64(64)},
		{"ToUint", func() (any, error) { return convertx.ToUint("42") }, uint(42)},
		{"ToUint8", func() (any, error) { return convertx.ToUint8("8") }, uint8(8)},
		{"ToUint16", func() (any, error) { return convertx.ToUint16("16") }, uint16(16)},
		{"ToUint32", func() (any, error) { return convertx.ToUint32("32") }, uint32(32)},
		{"ToUint64", func() (any, error) { return convertx.ToUint64("64") }, uint64(64)},
		{"ToFloat32", func() (any, error) { return convertx.ToFloat32("1.5") }, float32(1.5)},
		{"ToByte", func() (any, error) { return convertx.ToByte("65") }, byte(65)},
		{"ToRune", func() (any, error) { return convertx.ToRune("97") }, rune(97)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn()
			testx.NoError(t, err)
			testx.Equal(t, got, tt.want)
		})
	}
}

func TestStructConversion(t *testing.T) {
	type src struct {
		Name string
		Age  int
	}
	type dst struct {
		Name string
		Age  string
	}
	s := src{Name: "Alice", Age: 30}
	var d dst
	err := convertx.Struct(&s, &d)
	testx.NoError(t, err)
	testx.Equal(t, d.Name, "Alice")
	testx.Equal(t, d.Age, "30")

	// 泛型版本。
	d2, err := convertx.ToStruct[dst](&s)
	testx.NoError(t, err)
	testx.Equal(t, d2.Age, "30")

	// struct ↔ map。
	m, err := convertx.StructToMap(s)
	testx.NoError(t, err)
	testx.Equal(t, m["Name"], "Alice")

	var d3 dst
	err = convertx.MapToStruct(map[string]any{"Name": "Bob", "Age": "25"}, &d3)
	testx.NoError(t, err)
	testx.Equal(t, d3.Name, "Bob")
}

func TestCollectionConversion(t *testing.T) {
	// slice → slice。
	sl, err := convertx.SliceToSlice[string, int]([]string{"1", "2"})
	testx.NoError(t, err)
	testx.Equal(t, sl, []int{1, 2})

	// map → map。
	mp, err := convertx.MapToMap[string, string, string, int](map[string]string{"a": "1"})
	testx.NoError(t, err)
	testx.Equal(t, mp, map[string]int{"a": 1})

	// slice → map。
	sm, err := convertx.SliceToMap([]int{1, 2}, func(v int) int { return v })
	testx.NoError(t, err)
	testx.Equal(t, len(sm), 2)

	// map → slice。
	ms, err := convertx.MapToSlice(map[string]int{"a": 1})
	testx.NoError(t, err)
	testx.Equal(t, ms, []int{1})
}

func TestTimeConversion(t *testing.T) {
	tm, err := convertx.ToTime("2026-08-17T10:00:00Z")
	testx.NoError(t, err)
	testx.Equal(t, tm.Location(), time.UTC)

	s := convertx.TimeToString(tm, time.RFC3339)
	testx.Equal(t, s, "2026-08-17T10:00:00Z")

	ts := convertx.TimeToTimestamp(tm, convertx.UnitSecond)
	testx.Equal(t, ts, tm.Unix())

	back := convertx.TimestampToTime(ts, convertx.UnitSecond)
	testx.Equal(t, back.Unix(), ts)

	testx.Equal(t, convertx.ToUTC(tm).Location(), time.UTC)
	testx.Equal(t, convertx.ToLocal(tm).Location(), time.Local)
	testx.Equal(t, convertx.ToLocation(tm, time.UTC).Location(), time.UTC)

	d, err := convertx.ToDuration("1h30m")
	testx.NoError(t, err)
	testx.Equal(t, d, 90*time.Minute)
}

func TestValidation(t *testing.T) {
	v, err := convertx.ToAndValidate[int]("42", "min=0,max=100")
	testx.NoError(t, err)
	testx.Equal(t, v, 42)

	_, err = convertx.ToAndValidate[int]("150", "min=0,max=100")
	testx.Error(t, err)
	testx.ErrCode(t, err, convertx.ErrValidationFailed)

	testx.NoError(t, convertx.Validate("user@example.com", "email"))
	err = convertx.Validate("not-email", "email")
	testx.Error(t, err)
	testx.ErrCode(t, err, convertx.ErrValidationFailed)

	testx.NoError(t, convertx.ValidateField("abc", "min=2"))
}

func TestRegisterConverter(t *testing.T) {
	type status int
	err := convertx.RegisterConverter(func(s status) (string, error) {
		return "status", nil
	})
	testx.NoError(t, err)

	s, err := convertx.To[string](status(1))
	testx.NoError(t, err)
	testx.Equal(t, s, "status")
}

func TestErrorExtraction(t *testing.T) {
	_, err := convertx.ToInt("abc")
	var e *errx.Error
	testx.True(t, errors.As(err, &e))
	testx.Equal(t, e.Code(), convertx.ErrConversionFailed)
}

func TestOptionFunctions(t *testing.T) {
	// 根包选项函数应可公开调用。
	v, err := convertx.To[int]("abc", convertx.WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)

	v, err = convertx.To[int]("0", convertx.WithDefault(-1))
	testx.NoError(t, err)
	testx.Equal(t, v, -1)

	v, err = convertx.To[int]("abc", convertx.WithDefault("42"))
	testx.NoError(t, err)
	testx.Equal(t, v, 42)

	anyV, err := convertx.To[any](nil, convertx.WithDefault(1))
	testx.NoError(t, err)
	testx.Equal(t, anyV, 1)

	var p *int
	_, err = convertx.To[int](p, convertx.WithNilPointerAsError())
	testx.ErrCode(t, err, convertx.ErrNilPointer)

	tm, err := convertx.ToTime("2026/08/17 10:00:00", convertx.WithTimeLayout("2006/01/02 15:04:05"))
	testx.NoError(t, err)
	testx.Equal(t, tm.Year(), 2026)

	tm, err = convertx.ToTime("2026-08-17 10:00:00", convertx.WithDefaultLocation(time.UTC))
	testx.NoError(t, err)
	testx.Equal(t, tm.Location(), time.UTC)

	_, err = convertx.ToTime(1755484800, convertx.WithTimeUnit(convertx.UnitSecond))
	testx.NoError(t, err)

	_, err = convertx.To[int]("150", convertx.WithValidation("min=0,max=100"))
	testx.ErrCode(t, err, convertx.ErrValidationFailed)

	type tagSrc struct {
		UserName string `convertx:"username"`
	}
	type tagDst struct {
		Username string `convertx:"username"`
	}
	var tagOut tagDst
	err = convertx.Struct(&tagSrc{UserName: "u"}, &tagOut, convertx.WithTagName("convertx"))
	testx.NoError(t, err)
	testx.Equal(t, tagOut.Username, "u")

	type strictSrc struct {
		Extra string
	}
	type strictDst struct {
		Name string
	}
	err = convertx.Struct(&strictSrc{Extra: "x"}, &strictDst{}, convertx.WithStrictMode())
	testx.ErrCode(t, err, convertx.ErrUnmatchedField)
}
