package core

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

// ---------- error_util 分支 ----------

func TestWrapErrorNil(t *testing.T) {
	testx.Nil(t, wrapError(nil, ErrConversionFailed, nil, nil, "", nil))
}

func TestWrapErrorEmptyCode(t *testing.T) {
	err := wrapError(errors.New("x"), "", nil, nil, "", nil)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestNewErrorEmptyCode(t *testing.T) {
	err := newError("", "msg", nil, nil, "", nil)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestWithFieldPathBranches(t *testing.T) {
	// err 为 nil。
	testx.Nil(t, withFieldPath(nil, "X"))

	// fieldPath 为空：原样返回。
	plain := errors.New("plain")
	testx.True(t, withFieldPath(plain, "") == plain)

	// 非 errx 错误：原样返回。
	testx.True(t, withFieldPath(plain, "X") == plain)

	// errx 错误：附加 field_path。
	e := newError(ErrConversionFailed, "msg", nil, nil, "", nil)
	wrapped := withFieldPath(e, "A.B")
	var e2 *errx.Error
	testx.True(t, errors.As(wrapped, &e2))
	found := false
	for _, kv := range e2.Fields() {
		if kv.Key == "field_path" && kv.Value == "A.B" {
			found = true
		}
	}
	testx.True(t, found)
}

// ---------- scalar 边界 ----------

func TestFloat32StringOverflow(t *testing.T) {
	_, err := To[float32]("1e300")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrOverflow)
}

func TestInt32Uint32Boundaries(t *testing.T) {
	_, err := To[int32](2147483648)
	testx.ErrCode(t, err, ErrOverflow)

	_, err = To[uint32]("4294967296")
	testx.ErrCode(t, err, ErrOverflow)

	// uintptr 溢出。
	_, err = To[uintptr](^uintptr(0))
	testx.NoError(t, err)
}

func TestFloatToIntTargetRange(t *testing.T) {
	// float 值在 int64 范围内但超出 int8。
	_, err := To[int8](200.0)
	testx.ErrCode(t, err, ErrOverflow)
}

func TestFloatToUintNegative(t *testing.T) {
	_, err := To[uint](-1.5)
	testx.ErrCode(t, err, ErrOverflow)
}

// ---------- time 分支 ----------

func TestToTimeZuluSuffix(t *testing.T) {
	tm, err := ToTime("2026-08-17T10:00:00Z")
	testx.NoError(t, err)
	testx.Equal(t, tm.Location(), time.UTC)
}

func TestToTimeRFC822ZOffset(t *testing.T) {
	tm, err := ToTime("17 Aug 26 10:00 -0700")
	testx.NoError(t, err)
	testx.False(t, tm.IsZero())
	testx.Equal(t, tm.Format("-0700"), "-0700")
}

func TestTimeToTimestampMicrosecond(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	testx.Equal(t, TimeToTimestamp(tm, UnitMicrosecond), tm.UnixMicro())
}

func TestDurationToIntegerTargets(t *testing.T) {
	// duration → int8 溢出。
	_, err := To[int8](time.Hour)
	testx.ErrCode(t, err, ErrOverflow)

	// duration → uint。
	u, err := To[uint](time.Second)
	testx.NoError(t, err)
	testx.Equal(t, u, uint(time.Second))

	// duration → 不支持的类型。
	_, err = To[bool](time.Second)
	testx.Error(t, err)
}

func TestToInt64DirectCall(t *testing.T) {
	// 直接调用 toInt64 的非整数分支。
	_, err := toInt64(reflect.ValueOf("x"))
	testx.Error(t, err)
}

func TestToDurationUintOverflow(t *testing.T) {
	_, err := ToDuration(uint64(math.MaxInt64) + 100)
	testx.Error(t, err)
}

// ---------- struct/map 分支 ----------

func TestMapKeyTagShadowsFieldName(t *testing.T) {
	// 目标字段 tag 名与 key 不同且屏蔽字段名时，key 不匹配。
	type dst struct {
		Name string `convertx:"other"`
	}
	m := map[string]any{"Name": "x"}
	var out dst
	err := MapToStruct(m, &out)
	testx.NoError(t, err)
	testx.Equal(t, out.Name, "")
}

func TestMapToStructValueConversionFail(t *testing.T) {
	m := map[string]any{"Name": []int{1}}
	var dst structTarget
	err := MapToStruct(m, &dst)
	testx.Error(t, err)
}

func TestStructToMapStringTarget(t *testing.T) {
	// 目标 map[string]string：字段类型需转换为 string。
	type strSrc struct {
		Name string
		Age  int
	}
	src := strSrc{Name: "x", Age: 40}
	out, err := Convert(src, reflect.TypeOf(map[string]string{}))
	testx.NoError(t, err)
	testx.Equal(t, out, map[string]string{"Name": "x", "Age": "40"})

	// 无法转换为 string 的字段类型导致整体失败。
	type badSrc struct {
		Tags []string
	}
	src2 := badSrc{Tags: []string{"t"}}
	_, err = Convert(src2, reflect.TypeOf(map[string]string{}))
	testx.Error(t, err)
}

func TestEmbeddedNonStruct(t *testing.T) {
	// 匿名字段为非结构体类型（如命名 int）：作为普通字段处理。
	type EmbeddedInt int
	type embedMix struct {
		EmbeddedInt
		A string
	}
	src := embedMix{EmbeddedInt: 7, A: "x"}
	m, err := StructToMap(src)
	testx.NoError(t, err)
	testx.Equal(t, m["EmbeddedInt"], EmbeddedInt(7))
	testx.Equal(t, m["A"], "x")
}

func TestStructFieldNilInterface(t *testing.T) {
	// 字段级 nil 接口：默认模式转为目标零值。
	type srcT struct {
		V any
	}
	type dstT struct {
		V int
	}
	src := srcT{}
	var dst dstT
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.V, 0)

	// 字段级 nil 接口 + nilPointerAsError。
	err = Struct(&src, &dst, WithNilPointerAsError())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrNilPointer)
}

func TestStructFieldAnyTarget(t *testing.T) {
	// 目标字段为 any：源值原样放入。
	type srcT struct {
		V int
	}
	type dstT struct {
		V any
	}
	src := srcT{V: 42}
	var dst dstT
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.V, 42)
}

func TestMapToMapKeyConversionFail(t *testing.T) {
	// complex key 转 string 不支持，整体失败。
	src := map[complex64]int{1 + 2i: 3}
	_, err := MapToMap[complex64, int, string, int](src)
	testx.Error(t, err)
}

func TestConvertNonCollectionToSlice(t *testing.T) {
	// 源不是集合，目标为 slice。
	_, err := Convert("abc", reflect.TypeOf([]int{}))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

// ---------- 注册表断言失败分支 ----------

func TestRunConverterAssertionFailure(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	// 人为构造错配的注册表项：key 为 (string → int)，内部 fn 断言 int。
	globalRegistry.mu.Lock()
	globalRegistry.table[convKey{src: reflect.TypeOf(""), dst: reflect.TypeOf(0)}] = func(src any) (any, error) {
		return runConverter(func(s int) (int, error) { return s, nil }, src)
	}
	globalRegistry.mu.Unlock()

	_, err := To[int]("x")
	testx.Error(t, err)
}

// ---------- To 接口目标断言 ----------

type testIface interface {
	String() string
}

type testIfaceImpl struct{}

func (testIfaceImpl) String() string { return "impl" }

func TestToInterfaceTarget(t *testing.T) {
	// 未实现接口：断言失败。
	_, err := To[testIface]("x")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	// 实现接口：成功。
	v, err := To[testIface](testIfaceImpl{})
	testx.NoError(t, err)
	testx.Equal(t, v.String(), "impl")
}

func TestToAnyNil(t *testing.T) {
	v, err := To[any](nil)
	testx.NoError(t, err)
	testx.Nil(t, v)
}

// ---------- 覆盖率补齐：标量边界 ----------

func TestStringToInt8Overflow(t *testing.T) {
	// 字符串解析后的溢出检测。
	_, err := To[int8]("128")
	testx.ErrCode(t, err, ErrOverflow)
}

func TestInfToInt(t *testing.T) {
	_, err := To[int64](math.Inf(1))
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestNaNToUint(t *testing.T) {
	_, err := To[uint](math.NaN())
	testx.ErrCode(t, err, ErrOverflow)
}

func TestFloatToUintSuccess(t *testing.T) {
	v, err := To[uint](5.7)
	testx.NoError(t, err)
	testx.Equal(t, v, uint(5))
}

func TestUintptrRange(t *testing.T) {
	v, err := To[uintptr](uint64(123))
	testx.NoError(t, err)
	testx.Equal(t, v, uintptr(123))
}

// ---------- 覆盖率补齐：集合 ----------

func TestArrayTargetElementFailure(t *testing.T) {
	_, err := Convert([]string{"1", "bad"}, reflect.TypeOf([2]int{}))
	testx.Error(t, err)
}

// ---------- 覆盖率补齐：结构体 ----------

func TestTargetFieldOmitted(t *testing.T) {
	// 目标字段 tag "-"：匹配时跳过。
	type src struct {
		Skip string
	}
	type dst struct {
		Skip string `convertx:"-"`
	}
	srcVal := src{Skip: "x"}
	var dstVal dst
	err := Struct(&srcVal, &dstVal)
	testx.NoError(t, err)
	testx.Equal(t, dstVal.Skip, "")

	// map → struct 目标字段同样跳过。
	m := map[string]any{"Skip": "y"}
	err = MapToStruct(m, &dstVal)
	testx.NoError(t, err)
	testx.Equal(t, dstVal.Skip, "")
}

func TestPointerEmbeddedField(t *testing.T) {
	// 指针型匿名字段展开。
	type pEmbed struct {
		*embeddedBase
		Own string
	}
	src := pEmbed{embeddedBase: &embeddedBase{BaseField: "base"}, Own: "own"}
	m, err := StructToMap(src)
	testx.NoError(t, err)
	testx.Equal(t, m["BaseField"], "base")
	testx.Equal(t, m["Own"], "own")
}

// ---------- 覆盖率补齐：时间 ----------

func TestTimeToInt8Overflow(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	_, err := To[int8](tm)
	testx.ErrCode(t, err, ErrOverflow)
}

func TestTimeToUnsupportedTarget(t *testing.T) {
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	_, err := To[bool](tm)
	testx.Error(t, err)
}

// ---------- 覆盖率补齐：内部兜底分支 ----------

func TestParseNumericStringNonNumericTarget(t *testing.T) {
	strType := reflect.TypeOf("")
	_, err := parseNumericString("1", strType, strType, "")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestConvertNumberUnsupported(t *testing.T) {
	strType := reflect.TypeOf("")
	_, err := convertNumber(reflect.ValueOf("x"), strType, strType, "")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

// ---------- 覆盖率补齐：最终分支 ----------

func TestFloat32ToFloat64(t *testing.T) {
	v, err := To[float64](float32(1.5))
	testx.NoError(t, err)
	testx.Equal(t, v, float64(float32(1.5)))
}

func TestUintToUintOverflow(t *testing.T) {
	_, err := To[uint8](uint64(300))
	testx.ErrCode(t, err, ErrOverflow)
}

// ---------- 接口目标与指针方法集 ----------

type edgeReaderSrc struct {
	R *bytes.Buffer
}

type edgeReaderDst struct {
	R io.Reader
}

func TestStructInterfaceFieldAssignable(t *testing.T) {
	src := edgeReaderSrc{R: bytes.NewBufferString("x")}
	var dst edgeReaderDst
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.NotNil(t, dst.R)
}

func TestStructInterfaceFieldNotAssignable(t *testing.T) {
	type srcT struct {
		V int
	}
	type dstT struct {
		V testIface
	}
	src := srcT{V: 1}
	var dst dstT
	err := Struct(&src, &dst)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestToPointerInterfaceTarget(t *testing.T) {
	// *any 目标应构造接口指针，而不是 *int。
	v, err := To[*any](42)
	testx.NoError(t, err)
	testx.NotNil(t, v)
	testx.Equal(t, *v, 42)
}

// ---------- nil 指针嵌入字段 ----------

type edgePtrBase struct {
	Base string
}

type edgePtrEmbedSrc struct {
	*edgePtrBase
	Own string
}

type edgePtrEmbedDst struct {
	*edgePtrBase
	Own string
}

func TestNilPointerEmbeddedSource(t *testing.T) {
	src := edgePtrEmbedSrc{Own: "o"}
	m, err := StructToMap(src)
	testx.NoError(t, err)
	_, hasBase := m["Base"]
	testx.False(t, hasBase)
	testx.Equal(t, m["Own"], "o")
}

func TestNilPointerEmbeddedDestination(t *testing.T) {
	type srcT struct {
		Base string
		Own  string
	}
	src := srcT{Base: "b", Own: "o"}
	var dst edgePtrEmbedDst
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Nil(t, dst.edgePtrBase)
	testx.Equal(t, dst.Own, "o")
}

func TestMapToStructNilPointerEmbeddedDestination(t *testing.T) {
	var dst edgePtrEmbedDst
	err := MapToStruct(map[string]any{"Base": "b", "Own": "o"}, &dst)
	testx.NoError(t, err)
	testx.Nil(t, dst.edgePtrBase)
	testx.Equal(t, dst.Own, "o")
}

type EdgeExportedBase struct {
	Base string
}

type edgeExportedEmbedDst struct {
	*EdgeExportedBase
	Own string
}

func TestExportedNilPointerEmbeddedDestination(t *testing.T) {
	type srcT struct {
		Base string
		Own  string
	}
	src := srcT{Base: "b", Own: "o"}
	var dst edgeExportedEmbedDst
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.NotNil(t, dst.EdgeExportedBase)
	testx.Equal(t, dst.Base, "b")
	testx.Equal(t, dst.Own, "o")
}

func TestNilPointerEmbeddedSourceStrictError(t *testing.T) {
	src := edgePtrEmbedSrc{Own: "o"}
	_, err := StructToMap(src, WithNilPointerAsError())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrNilPointer)
}

// ---------- StructToMap 与 time.Time ----------

func TestStructToMapTimeField(t *testing.T) {
	type srcT struct {
		Name string
		Time time.Time
	}
	tm := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	src := srcT{Name: "x", Time: tm}
	m, err := StructToMap(src)
	testx.NoError(t, err)
	testx.Equal(t, m["Name"], "x")
	got, ok := m["Time"].(time.Time)
	testx.True(t, ok)
	testx.True(t, got.Equal(tm))
}
