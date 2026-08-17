package core

import (
	"testing"

	"github.com/lcylpzls/testx"
)

func TestConvertNilSource(t *testing.T) {
	// 源为 nil 且目标非指针：零值 + 错误。
	_, err := To[int](nil)
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)

	// 源为 nil 且目标为指针：nil 指针，无错误。
	p, err := To[*int](nil)
	testx.NoError(t, err)
	testx.Nil(t, p)
}

func TestConvertNilPointer(t *testing.T) {
	// nil *int → int：默认返回零值。
	var p *int
	v, err := To[int](p)
	testx.NoError(t, err)
	testx.Equal(t, v, 0)

	// nil *int → *int64：默认返回 nil 指针。
	pp, err := To[*int64](p)
	testx.NoError(t, err)
	testx.Nil(t, pp)

	// WithNilPointerAsError：返回错误。
	_, err = To[int](p, WithNilPointerAsError())
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrNilPointer)
}

func TestConvertPointerDeref(t *testing.T) {
	x := 42
	v, err := To[int](&x)
	testx.NoError(t, err)
	testx.Equal(t, v, 42)

	// 多级指针。
	y := 7
	pp := &y
	ppp := &pp
	v, err = To[int](ppp)
	testx.NoError(t, err)
	testx.Equal(t, v, 7)
}

func TestConvertPointerRebuild(t *testing.T) {
	// int → *int：取地址，不共享源地址。
	v, err := To[*int](42)
	testx.NoError(t, err)
	testx.NotNil(t, v)
	testx.Equal(t, *v, 42)

	// int → **int：两层取地址。
	pp, err := To[**int](42)
	testx.NoError(t, err)
	testx.NotNil(t, pp)
	testx.Equal(t, **pp, 42)
}

func TestConvertPointerToDifferentPointer(t *testing.T) {
	x := int64(100)
	// *int64 → *int8：解引用 → 基础转换 → 取地址。
	v, err := To[*int8](&x)
	testx.NoError(t, err)
	testx.NotNil(t, v)
	testx.Equal(t, *v, int8(100))

	// 溢出场景：*int64 → *int8。
	big := int64(1000)
	_, err = To[*int8](&big)
	testx.ErrCode(t, err, ErrOverflow)
}

func TestConvertTypeMatch(t *testing.T) {
	// 相同类型直接返回。
	v, err := To[string]("hello")
	testx.NoError(t, err)
	testx.Equal(t, v, "hello")

	// *int → *int：值复制重建（PRD 4.6.3：取地址操作返回新分配的指针，不共享源地址）。
	x := 99
	p1 := &x
	p2, err := To[*int](p1)
	testx.NoError(t, err)
	testx.NotNil(t, p2)
	testx.Equal(t, *p2, 99)
	testx.False(t, p1 == p2)
}

func TestConvertToAny(t *testing.T) {
	// 目标为 any：直接返回源值。
	v, err := To[any]("keep")
	testx.NoError(t, err)
	testx.Equal(t, v, "keep")
}

func TestConvertInterfaceSource(t *testing.T) {
	// 接口字段/值：解包后转换。
	var i any = "123"
	v, err := To[int](i)
	testx.NoError(t, err)
	testx.Equal(t, v, 123)
}

func TestConvertUnsupportedTarget(t *testing.T) {
	// chan 目标类型不支持。
	_, err := To[chan int]("x")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestConvertConcurrentSafety(t *testing.T) {
	// 并发调用不 panic、不产生数据竞争。
	testx.Concurrently(t, 32, func() {
		_, _ = To[int]("42")
		_, _ = To[string](123)
	})
}
