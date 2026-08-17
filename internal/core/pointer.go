package core

import "reflect"

// buildPointer 按 depth 层取地址重建指针链，返回新分配的指针。
func buildPointer(v reflect.Value, depth int) reflect.Value {
	for i := 0; i < depth; i++ {
		p := reflect.New(v.Type())
		p.Elem().Set(v)
		v = p
	}
	return v
}

// hasNilPointer 检查反射值链（指针/接口逐层解引用）中是否存在 nil。
// 用于顶层转换的 nil 语义判断；map/slice 的 nil 由集合转换自行处理。
func hasNilPointer(v reflect.Value) bool {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	return false
}
