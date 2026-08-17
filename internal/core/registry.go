package core

import (
	"fmt"
	"reflect"
	"sync"
)

// converterFunc 是统一的自定义转换器执行形态：输入源值，输出目标值与错误。
type converterFunc func(src any) (any, error)

// convKey 是转换器注册表的唯一键：源类型 + 目标类型。
type convKey struct {
	src reflect.Type
	dst reflect.Type
}

// registry 是全局自定义转换器注册表，并发读安全。
type registry struct {
	mu    sync.RWMutex
	table map[convKey]converterFunc
}

// globalRegistry 是包级全局注册表，init 阶段注册后运行时只读。
var globalRegistry = &registry{table: make(map[convKey]converterFunc)}

// RegisterConverter 注册自定义类型转换器，全局生效，优先级高于内置转换器。
// 同一 (源类型, 目标类型) 重复注册返回 ErrDuplicateConverter。
// 建议在 init() 或程序启动阶段注册；注册后不可注销、不可覆盖。
func RegisterConverter[Src any, Dst any](fn func(Src) (Dst, error)) error {
	var s Src
	var d Dst
	srcType := reflect.TypeOf(s)
	dstType := reflect.TypeOf(d)
	if srcType == nil || dstType == nil {
		return newError(ErrConversionFailed, "转换器源类型与目标类型均不能为接口类型", nil, nil, "", nil)
	}
	key := convKey{src: srcType, dst: dstType}

	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if _, ok := globalRegistry.table[key]; ok {
		return newError(ErrDuplicateConverter,
			fmt.Sprintf("转换器已注册：%s → %s", srcType, dstType), srcType, dstType, "", nil)
	}
	globalRegistry.table[key] = func(src any) (any, error) {
		return runConverter(fn, src)
	}
	return nil
}

// runConverter 执行自定义转换器：类型断言、防御 panic、透传错误。
func runConverter[Src, Dst any](fn func(Src) (Dst, error), src any) (out Dst, err error) {
	defer func() {
		if r := recover(); r != nil {
			out = *new(Dst)
			err = newError(ErrConversionFailed,
				fmt.Sprintf("自定义转换器 panic：%v", r), nil, nil, "", src)
		}
	}()
	s, ok := src.(Src)
	if !ok {
		return *new(Dst), newError(ErrConversionFailed, "自定义转换器源类型断言失败", nil, nil, "", src)
	}
	return fn(s)
}

// lookupConverter 按 (源类型, 目标类型) 查找自定义转换器。
func lookupConverter(src, dst reflect.Type) (converterFunc, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	fn, ok := globalRegistry.table[convKey{src: src, dst: dst}]
	return fn, ok
}
