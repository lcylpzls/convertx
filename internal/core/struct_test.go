package core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

// structSource / structTarget 用于结构体转换测试。
type structSource struct {
	Name string
	Age  int
	Tags []string
}

type structTarget struct {
	Name string
	Age  string
	Tags []string
}

// tagSource / tagTarget 用于 tag 映射测试。
type tagSource struct {
	UserName string `convertx:"username"`
	SkipMe   string `convertx:"-"`
	Keep     string
}

type tagTarget struct {
	Username string `convertx:"username"`
	Keep     string
	Extra    string
}

// nestSource / nestTarget 用于嵌套结构体测试。
type nestSource struct {
	ID      int
	Address struct {
		City string
	}
}

type nestTarget struct {
	ID      int
	Address struct {
		City string
	}
}

// embeddedSource / embeddedTarget 用于匿名字段测试。
type embeddedBase struct {
	BaseField string
}

type embeddedSource struct {
	embeddedBase
	Own string
}

type embeddedTarget struct {
	embeddedBase
	Own string
}

// cycleSrc / cycleDst 用于循环引用测试（异类型触发递归转换）。
type cycleSrc struct {
	Value int
	Next  *cycleSrc
}

type cycleDst struct {
	Value int
	Next  *cycleDst
}

// mutualA / mutualB 用于匿名互嵌指针的无限递归回归测试。
type mutualA struct {
	Name string
	*mutualB
}

type mutualB struct {
	*mutualA
}

func TestStructBasic(t *testing.T) {
	src := structSource{Name: "Alice", Age: 30, Tags: []string{"a", "b"}}
	var dst structTarget
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "Alice")
	testx.Equal(t, dst.Age, "30")
	testx.Equal(t, dst.Tags, []string{"a", "b"})
}

func TestToStructGeneric(t *testing.T) {
	src := structSource{Name: "Bob", Age: 25}
	dst, err := ToStruct[structTarget](&src)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "Bob")
	testx.Equal(t, dst.Age, "25")
}

func TestStructTagMapping(t *testing.T) {
	src := tagSource{UserName: "carol", SkipMe: "hidden", Keep: "keep-me"}
	var dst tagTarget
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	// tag 映射优先级高于字段名。
	testx.Equal(t, dst.Username, "carol")
	// 忽略字段不复制。
	testx.Equal(t, dst.Keep, "keep-me")
	// 未匹配字段（Extra）默认忽略不报错。
	testx.Equal(t, dst.Extra, "")
}

func TestStructStrictMode(t *testing.T) {
	// 源字段 Extra 在目标中不存在，严格模式报错。
	type strictSource struct {
		Name  string
		Extra string
	}
	src := strictSource{Name: "x", Extra: "y"}
	var dst structTarget
	err := Struct(&src, &dst, WithStrictMode())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrUnmatchedField)
}

func TestStructNested(t *testing.T) {
	src := nestSource{ID: 1}
	src.Address.City = "Beijing"
	var dst nestTarget
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.ID, 1)
	testx.Equal(t, dst.Address.City, "Beijing")
}

func TestStructEmbedded(t *testing.T) {
	src := embeddedSource{embeddedBase: embeddedBase{BaseField: "base"}, Own: "own"}
	var dst embeddedTarget
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.BaseField, "base")
	testx.Equal(t, dst.Own, "own")
}

func TestStructCircularReference(t *testing.T) {
	n1 := &cycleSrc{Value: 1}
	n2 := &cycleSrc{Value: 2}
	n1.Next = n2
	n2.Next = n1 // 形成环

	var dst cycleDst
	err := Struct(n1, &dst)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrCircularReference)
}

func TestStructNilSource(t *testing.T) {
	var dst structTarget
	// 源为 nil：目标保持零值，无错误。
	err := Struct(nil, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst, structTarget{})

	// 源为 nil 指针。
	var src *structSource
	err = Struct(src, &dst)
	testx.NoError(t, err)

	// WithNilPointerAsError。
	err = Struct(src, &dst, WithNilPointerAsError())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrNilPointer)
}

func TestStructInvalidTarget(t *testing.T) {
	src := structSource{}
	// 目标为 nil。
	err := Struct(&src, nil)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	// 目标非指针。
	var dst structTarget
	err = Struct(&src, dst)
	testx.Error(t, err)

	// 目标不是结构体指针。
	var n int
	err = Struct(&src, &n)
	testx.Error(t, err)

	// 目标为 nil 指针。
	var dstPtr *structTarget
	err = Struct(&src, dstPtr)
	testx.Error(t, err)
}

func TestStructFieldPathError(t *testing.T) {
	// 嵌套结构体字段转换失败时携带 field_path。
	type inner struct {
		Num string
	}
	type outer struct {
		Inner inner
	}
	type target struct {
		Inner struct {
			Num int
		}
	}
	src := outer{Inner: inner{Num: "abc"}}
	var dst target
	err := Struct(&src, &dst)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	var e *errx.Error
	testx.True(t, errors.As(err, &e))
	found := false
	for _, kv := range e.Fields() {
		if kv.Key == "field_path" && kv.Value == "Inner.Num" {
			found = true
		}
	}
	testx.True(t, found)
}

func TestStructCaseSensitive(t *testing.T) {
	// 字段名大小写敏感：name 不匹配 Name。
	type srcT struct {
		name string
	}
	type dstT struct {
		Name string
	}
	src := srcT{name: "x"}
	var dst dstT
	err := Struct(&src, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "")
}

func TestStructSourceNotStruct(t *testing.T) {
	var dst structTarget
	err := Struct(42, &dst)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestStructToMap(t *testing.T) {
	src := tagSource{UserName: "dave", SkipMe: "hidden", Keep: "keep"}
	m, err := StructToMap(&src)
	testx.NoError(t, err)
	// tag 名作为 map key。
	testx.Equal(t, m["username"], "dave")
	// 忽略字段不导出。
	_, hasSkip := m["SkipMe"]
	testx.False(t, hasSkip)
	testx.Equal(t, m["Keep"], "keep")
}

func TestStructToMapNested(t *testing.T) {
	src := nestSource{ID: 7}
	src.Address.City = "Shanghai"
	m, err := StructToMap(src)
	testx.NoError(t, err)
	testx.Equal(t, m["ID"], 7)
	addr, ok := m["Address"].(map[string]any)
	testx.True(t, ok)
	testx.Equal(t, addr["City"], "Shanghai")
}

func TestMapToStruct(t *testing.T) {
	m := map[string]any{"Name": "eve", "Age": "28", "Tags": []string{"x"}}
	var dst structTarget
	err := MapToStruct(m, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "eve")
	testx.Equal(t, dst.Age, "28")
	testx.Equal(t, dst.Tags, []string{"x"})
}

func TestMapToStructStrongTyped(t *testing.T) {
	m := map[string]string{"Name": "frank", "Age": "33"}
	var dst structTarget
	err := MapToStruct(m, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "frank")
	testx.Equal(t, dst.Age, "33")
}

func TestMapToStructStrictMode(t *testing.T) {
	m := map[string]any{"Name": "x", "UnknownKey": 1}
	var dst structTarget
	err := MapToStruct(m, &dst, WithStrictMode())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrUnmatchedField)

	// 默认模式忽略多余 key。
	err = MapToStruct(m, &dst)
	testx.NoError(t, err)
}

func TestMapToStructNil(t *testing.T) {
	var dst structTarget
	err := MapToStruct(nil, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst, structTarget{})

	// nil map（接口非 nil）。
	var m map[string]any
	err = MapToStruct(m, &dst)
	testx.NoError(t, err)
}

func TestMapToStructTagKey(t *testing.T) {
	m := map[string]any{"username": "grace"}
	var dst tagTarget
	err := MapToStruct(m, &dst)
	testx.NoError(t, err)
	testx.Equal(t, dst.Username, "grace")
}

// TestStructSelfEmbeddingNoStackOverflow 验证匿名自嵌入指针类型不再导致无限递归。
func TestStructSelfEmbeddingNoStackOverflow(t *testing.T) {
	type node struct {
		Value int
		*node
	}
	type flat struct {
		Value int
	}
	// 自引用实例：内嵌 *node 指向自身（元数据解析阶段即触发环检测）。
	src := node{Value: 1}
	src.node = &src
	dst, err := To[flat](src)
	testx.NoError(t, err)
	testx.Equal(t, dst.Value, 1)
}

// TestStructMutualEmbeddingNoStackOverflow 验证匿名互嵌指针类型不再导致无限递归。
// 互嵌类型无法用局部类型表达（局部类型声明不能前向引用），因此提升为包级。
func TestStructMutualEmbeddingNoStackOverflow(t *testing.T) {
	type flat struct {
		Name string
	}
	src := mutualA{Name: "x", mutualB: &mutualB{}}
	dst, err := To[flat](src)
	testx.NoError(t, err)
	testx.Equal(t, dst.Name, "x")
}

// TestStructNilPointerFieldConversion 验证 nil 指针字段跨类型转换不再 panic。
func TestStructNilPointerFieldConversion(t *testing.T) {
	type s1 struct{ V int }
	type s2 struct{ V string }
	type srcT struct{ P *s1 }
	type dstT struct{ P *s2 }
	var d dstT
	err := Struct(srcT{}, &d)
	testx.NoError(t, err)
	testx.Equal(t, d.P, (*s2)(nil))
}

// TestConvertNilPointerInCollections 验证 slice/map/接口中的 nil 指针转换不再 panic。
func TestConvertNilPointerInCollections(t *testing.T) {
	type s1 struct{ V int }
	type s2 struct{ V string }

	sl, err := To[[]*s2]([]*s1{nil})
	testx.NoError(t, err)
	testx.Equal(t, sl, []*s2{nil})

	m, err := To[map[string]*s2](map[string]*s1{"k": nil})
	testx.NoError(t, err)
	testx.Equal(t, m, map[string]*s2{"k": nil})

	type srcI struct{ V any }
	type dstI struct{ V *int }
	var d dstI
	err = Struct(srcI{}, &d)
	testx.NoError(t, err)
	testx.Equal(t, d.V, (*int)(nil))
}

// TestParseStructMetaCache 验证元信息缓存：同键命中同一实例，不同 tag 名独立。
func TestParseStructMetaCache(t *testing.T) {
	f1 := parseStructMeta(reflect.TypeOf(structSource{}), "convertx")
	f2 := parseStructMeta(reflect.TypeOf(structSource{}), "convertx")
	testx.Equal(t, f1, f2)
	// 同键返回同一缓存实例（共享底层数组）。
	testx.True(t, &f1[0] == &f2[0])

	// 不同 tag 名独立缓存，解析结果不同。
	g1 := parseStructMeta(reflect.TypeOf(tagSource{}), "convertx")
	testx.Equal(t, g1[0].TagName, "username")
	g2 := parseStructMeta(reflect.TypeOf(tagSource{}), "other")
	testx.Equal(t, g2[0].TagName, "")
}

// TestParseStructMetaCacheConcurrent 验证缓存并发读写安全（配合 -race）。
func TestParseStructMetaCacheConcurrent(t *testing.T) {
	testx.Concurrently(t, 32, func() {
		fields := parseStructMeta(reflect.TypeOf(structSource{}), "convertx")
		testx.Equal(t, len(fields), 3)
		_ = parseStructMeta(reflect.TypeOf(tagSource{}), "other")
	})
}
