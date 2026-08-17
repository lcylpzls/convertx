package core

import (
	"testing"

	"github.com/lcylpzls/testx"
)

func TestStructToMapBasic(t *testing.T) {
	src := structSource{Name: "henry", Age: 40, Tags: []string{"t1"}}
	m, err := StructToMap(src)
	testx.NoError(t, err)
	testx.Equal(t, m["Name"], "henry")
	testx.Equal(t, m["Age"], 40)
	testx.Equal(t, m["Tags"], []string{"t1"})
}

func TestStructToMapNonStructSource(t *testing.T) {
	_, err := StructToMap(42)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestStructToMapNonStringKeyTarget(t *testing.T) {
	// 通过内部 convert 触发：目标 map key 非 string。
	src := structSource{Name: "x"}
	_, err := Convert(src, reflectMapIntAnyType())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestMapToStructNonMapSource(t *testing.T) {
	var dst structTarget
	err := MapToStruct("not-map", &dst)
	testx.Error(t, err)
}

func TestMapToStructUnmatchedDefault(t *testing.T) {
	m := map[string]any{"Name": "x", "Missing": true}
	var dst structTarget
	err := MapToStruct(m, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "x")
}
