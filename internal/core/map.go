package core

import (
	"fmt"
	"reflect"
)

// convertMap 将源值转换为目标 map 类型。
// 源为结构体时执行 struct → map；源为 map 时执行 map → map（集合转换）。
func convertMap(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	switch srcVal.Kind() {
	case reflect.Struct:
		return convertStructToMap(srcVal, dstType, cfg, fieldPath)
	case reflect.Map:
		return convertMapToMap(srcVal, dstType, cfg, fieldPath)
	}
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("无法将 %s 转换为 map", srcVal.Type()),
		srcVal.Type(), dstType, fieldPath, srcVal.Interface())
}
