package core

import (
	"fmt"
	"reflect"
)

// convertSlice 将源 slice/array 转换为目标 slice/array 类型。
// 任一元素转换失败则整体失败；nil slice 返回 nil slice。
func convertSlice(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	srcKind := srcVal.Kind()
	if srcKind != reflect.Slice && srcKind != reflect.Array {
		return reflect.Value{}, newError(ErrConversionFailed,
			fmt.Sprintf("源类型 %s 不是 slice/array", srcVal.Type()),
			srcVal.Type(), dstType, fieldPath, srcVal.Interface())
	}
	// nil slice → nil slice（C04：nil 输入返回 nil 集合）。
	if srcKind == reflect.Slice && srcVal.IsNil() {
		return reflect.Zero(dstType), nil
	}

	// 目标为 array：长度必须一致，逐元素赋值。
	if dstType.Kind() == reflect.Array {
		if srcVal.Len() != dstType.Len() {
			return reflect.Value{}, newError(ErrConversionFailed,
				fmt.Sprintf("数组长度不匹配：源 %d，目标 %d", srcVal.Len(), dstType.Len()),
				srcVal.Type(), dstType, fieldPath, srcVal.Interface())
		}
		out := reflect.New(dstType).Elem()
		for i := 0; i < dstType.Len(); i++ {
			ev, err := convert(srcVal.Index(i), dstType.Elem(), cfg, joinFieldPath(fieldPath, fmt.Sprintf("[%d]", i)))
			if err != nil {
				return reflect.Value{}, err
			}
			out.Index(i).Set(ev)
		}
		return out, nil
	}

	// 目标为 slice：预分配容量，逐元素转换，保留原顺序。
	n := srcVal.Len()
	out := reflect.MakeSlice(reflect.SliceOf(dstType.Elem()), n, n)
	for i := 0; i < n; i++ {
		ev, err := convert(srcVal.Index(i), dstType.Elem(), cfg, joinFieldPath(fieldPath, fmt.Sprintf("[%d]", i)))
		if err != nil {
			return reflect.Value{}, err
		}
		out.Index(i).Set(ev)
	}
	return out, nil
}

// convertMapToMap 将源 map 转换为目标 map 类型（key/value 均可异类型转换）。
// 任一 key/value 转换失败则整体失败；nil map 返回 nil map。
func convertMapToMap(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	if srcVal.IsNil() {
		return reflect.Zero(dstType), nil
	}

	out := reflect.MakeMapWithSize(dstType, srcVal.Len())
	iter := srcVal.MapRange()
	for iter.Next() {
		nk, err := convert(iter.Key(), dstType.Key(), cfg, joinFieldPath(fieldPath, "[key]"))
		if err != nil {
			return reflect.Value{}, err
		}
		nv, err := convert(iter.Value(), dstType.Elem(), cfg, joinFieldPath(fieldPath, "[value]"))
		if err != nil {
			return reflect.Value{}, err
		}
		out.SetMapIndex(nk, nv)
	}
	return out, nil
}
