package core

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/lcylpzls/testx"
)

// reflectMapIntAnyType 构造 map[int]any 类型（key 非 string，用于错误路径测试）。
func reflectMapIntAnyType() reflect.Type {
	return reflect.TypeOf(map[int]any{})
}

func TestSliceToSliceBasic(t *testing.T) {
	src := []string{"1", "2", "3"}
	dst, err := SliceToSlice[string, int](src)
	testx.NoError(t, err)
	testx.Equal(t, dst, []int{1, 2, 3})

	// 相同类型。
	dst2, err := SliceToSlice[string, string](src)
	testx.NoError(t, err)
	testx.Equal(t, dst2, src)
}

func TestSliceToSliceNil(t *testing.T) {
	var src []string
	dst, err := SliceToSlice[string, int](src)
	testx.NoError(t, err)
	testx.Nil(t, dst)
}

func TestSliceToSliceElementFailure(t *testing.T) {
	src := []string{"1", "abc", "3"}
	_, err := SliceToSlice[string, int](src)
	testx.Error(t, err)
	// 整体失败，无部分结果。
}

func TestSliceToSliceArrayTarget(t *testing.T) {
	src := []string{"1", "2"}
	out, err := Convert(src, reflect.TypeOf([2]int{}))
	testx.NoError(t, err)
	testx.Equal(t, out, [2]int{1, 2})

	// 长度不匹配。
	_, err = Convert(src, reflect.TypeOf([3]int{}))
	testx.Error(t, err)
}

func TestSliceToSliceNonSliceSource(t *testing.T) {
	_, err := SliceToSlice[string, int](nil)
	// nil slice 作为源：返回 nil，无错误。
	testx.NoError(t, err)
}

func TestMapToMapBasic(t *testing.T) {
	src := map[string]string{"a": "1", "b": "2"}
	dst, err := MapToMap[string, string, string, int](src)
	testx.NoError(t, err)
	testx.Equal(t, dst, map[string]int{"a": 1, "b": 2})
}

func TestMapToMapKeyConversion(t *testing.T) {
	src := map[int]int{1: 10, 2: 20}
	dst, err := MapToMap[int, int, string, int64](src)
	testx.NoError(t, err)
	testx.Equal(t, dst, map[string]int64{"1": 10, "2": 20})
}

func TestMapToMapNil(t *testing.T) {
	var src map[string]string
	dst, err := MapToMap[string, string, string, int](src)
	testx.NoError(t, err)
	testx.Nil(t, dst)
}

func TestMapToMapElementFailure(t *testing.T) {
	src := map[string]string{"a": "1", "b": "bad"}
	_, err := MapToMap[string, string, string, int](src)
	testx.Error(t, err)
}

func TestSliceToMap(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}
	items := []item{{1, "a"}, {2, "b"}}
	m, err := SliceToMap(items, func(it item) int { return it.ID })
	testx.NoError(t, err)
	testx.Equal(t, len(m), 2)
	testx.Equal(t, m[1], item{1, "a"})
	testx.Equal(t, m[2], item{2, "b"})
}

func TestSliceToMapDuplicateKey(t *testing.T) {
	src := []int{1, 1, 2}
	m, err := SliceToMap(src, func(v int) int { return v })
	testx.NoError(t, err)
	// 重复 key 后者覆盖前者。
	testx.Equal(t, len(m), 2)
	testx.Equal(t, m[1], 1)
}

func TestSliceToMapKeyFnPanic(t *testing.T) {
	src := []int{1, 2}
	_, err := SliceToMap(src, func(v int) int {
		if v == 2 {
			panic("keyFn 崩溃")
		}
		return v
	})
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestMapToSlice(t *testing.T) {
	src := map[string]int{"a": 1, "b": 2}
	dst, err := MapToSlice(src)
	testx.NoError(t, err)
	testx.Equal(t, len(dst), 2)
	// 元素集合一致（顺序不保证）。
	testx.ElementsMatch(t, dst, []int{1, 2})
}

func TestCollectionFieldPath(t *testing.T) {
	src := []string{"ok", "bad"}
	_, err := SliceToSlice[string, int](src)
	testx.Error(t, err)
}

func TestConvertSliceNilCollection(t *testing.T) {
	// nil slice → nil 集合（C04）。
	var src []int
	out, err := Convert(src, reflect.TypeOf([]int64{}))
	testx.NoError(t, err)
	testx.Nil(t, out)
}

func TestConvertMapNilCollection(t *testing.T) {
	var src map[string]int
	out, err := Convert(src, reflect.TypeOf(map[string]int64{}))
	testx.NoError(t, err)
	testx.Nil(t, out)
}

func TestSliceToSliceAnyTarget(t *testing.T) {
	src := []int{1, 2}
	dst, err := SliceToSlice[int, any](src)
	testx.NoError(t, err)
	testx.Equal(t, dst, []any{1, 2})
}

func TestMapToMapAnyTarget(t *testing.T) {
	src := map[string]int{"a": 1}
	dst, err := MapToMap[string, int, string, any](src)
	testx.NoError(t, err)
	testx.Equal(t, dst, map[string]any{"a": 1})
}

func TestSliceToSliceInterfaceTarget(t *testing.T) {
	buf := bytes.NewBufferString("x")
	src := []*bytes.Buffer{buf}
	dst, err := SliceToSlice[*bytes.Buffer, io.Reader](src)
	testx.NoError(t, err)
	testx.Equal(t, len(dst), 1)
	var _ io.Reader = dst[0]
}
