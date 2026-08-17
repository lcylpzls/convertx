package core

import (
	"fmt"
	"reflect"
)

// convertStruct 将源值转换为目标结构体类型。
// 源可为结构体（struct → struct）或 map（map → struct）。
func convertStruct(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	if srcVal.Kind() == reflect.Map {
		return convertMapToStruct(srcVal, dstType, cfg, fieldPath)
	}
	if srcVal.Kind() != reflect.Struct {
		return reflect.Value{}, newError(ErrConversionFailed,
			fmt.Sprintf("源类型 %s 不是结构体或 map", srcVal.Type()),
			srcVal.Type(), dstType, fieldPath, srcVal.Interface())
	}
	srcType := srcVal.Type()

	srcFields := parseStructMeta(srcType, cfg.tagName)
	dstFields := parseStructMeta(dstType, cfg.tagName)

	out := reflect.New(dstType).Elem()
	var unmatched []string

	for _, sf := range srcFields {
		if sf.Omit {
			continue
		}
		df, ok := findFieldByName(dstFields, sf)
		if !ok {
			if cfg.strictMode {
				unmatched = append(unmatched, sf.Name)
			}
			continue
		}
		srcFv, ok := fieldValueByIndex(srcVal, sf.Index)
		if !ok {
			if cfg.nilPointerAsError {
				return reflect.Value{}, newError(ErrNilPointer, "源值为 nil 嵌入指针",
					srcVal.Type(), df.Type, joinFieldPath(fieldPath, df.Name), nil)
			}
			continue
		}
		dstFv, ok := fieldValueForSet(out, df.Index)
		if !ok {
			continue
		}
		// 字段类型相同 → 直接赋值（值复制）。
		if srcFv.Type() == dstFv.Type() {
			dstFv.Set(srcFv)
			continue
		}
		// 字段类型不同 → 递归转换（携带 field_path）。
		cv, err := convert(srcFv, dstFv.Type(), cfg, joinFieldPath(fieldPath, df.Name))
		if err != nil {
			return reflect.Value{}, err
		}
		dstFv.Set(cv)
	}

	if len(unmatched) > 0 {
		return reflect.Value{}, newError(ErrUnmatchedField,
			fmt.Sprintf("严格模式：存在未匹配字段 %v", unmatched),
			srcType, dstType, fieldPath, srcVal.Interface())
	}
	return out, nil
}

// findFieldByName 在目标字段列表中查找与源字段匹配的目标字段。
// 匹配优先级：源 tag 名 → 目标 tag 名 → 字段名（大小写敏感）。
func findFieldByName(dstFields []FieldMeta, sf FieldMeta) (FieldMeta, bool) {
	if sf.TagName != "" {
		if df, ok := matchField(dstFields, sf.TagName); ok {
			return df, true
		}
	}
	return matchField(dstFields, sf.Name)
}

// matchField 在字段列表中按名称匹配（优先 tag 名，其次字段名）。
func matchField(fields []FieldMeta, name string) (FieldMeta, bool) {
	for _, f := range fields {
		if f.Omit {
			continue
		}
		if f.TagName != "" && f.TagName == name {
			return f, true
		}
		if f.TagName == "" && f.Name == name {
			return f, true
		}
	}
	return FieldMeta{}, false
}

// convertMapToStruct 将 map 键值对填充到目标结构体字段。
// 支持强类型 map（如 map[string]string），key 转字符串后按 tag 名/字段名匹配。
func convertMapToStruct(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	dstFields := parseStructMeta(dstType, cfg.tagName)

	out := reflect.New(dstType).Elem()
	var unmatched []string

	iter := srcVal.MapRange()
	for iter.Next() {
		key := fmt.Sprint(iter.Key().Interface())
		df, ok := matchField(dstFields, key)
		if !ok {
			if cfg.strictMode {
				unmatched = append(unmatched, key)
			}
			continue
		}
		dstFv, ok := fieldValueForSet(out, df.Index)
		if !ok {
			continue
		}
		cv, err := convert(iter.Value(), dstFv.Type(), cfg, joinFieldPath(fieldPath, df.Name))
		if err != nil {
			return reflect.Value{}, err
		}
		dstFv.Set(cv)
	}

	if len(unmatched) > 0 {
		return reflect.Value{}, newError(ErrUnmatchedField,
			fmt.Sprintf("严格模式：存在未匹配字段 %v", unmatched),
			srcVal.Type(), dstType, fieldPath, srcVal.Interface())
	}
	return out, nil
}

// mapAnyType 是 map[string]any 的反射类型，用于嵌套结构体转嵌套 map。
var mapAnyType = reflect.TypeOf(map[string]any{})

// convertStructToMap 将源结构体导出为目标 map 类型。
// map key 默认使用字段名，tag 指定时使用 tag 值；嵌套结构体递归转为嵌套 map。
func convertStructToMap(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	if dstType.Key().Kind() != reflect.String {
		return reflect.Value{}, newError(ErrConversionFailed,
			fmt.Sprintf("struct → map 的目标 map key 必须是 string，当前为 %s", dstType.Key()),
			srcVal.Type(), dstType, fieldPath, srcVal.Interface())
	}

	srcFields := parseStructMeta(srcVal.Type(), cfg.tagName)

	out := reflect.MakeMapWithSize(dstType, len(srcFields))
	for _, sf := range srcFields {
		if sf.Omit {
			continue
		}
		key := sf.Name
		if sf.TagName != "" {
			key = sf.TagName
		}
		fv, ok := fieldValueByIndex(srcVal, sf.Index)
		if !ok {
			if cfg.nilPointerAsError {
				return reflect.Value{}, newError(ErrNilPointer, "源值为 nil 嵌入指针",
					srcVal.Type(), sf.Type, joinFieldPath(fieldPath, key), nil)
			}
			continue
		}
		elemType := dstType.Elem()
		var cv reflect.Value
		var err error
		subPath := joinFieldPath(fieldPath, key)
		if elemType.Kind() == reflect.Interface {
			// 目标元素为 any：嵌套结构体转为嵌套 map，其余值原样保留。
			if fv.Type() == timeType {
				cv = fv
			} else if fv.Kind() == reflect.Struct {
				cv, err = convert(fv, mapAnyType, cfg, subPath)
			} else {
				cv = fv
			}
		} else {
			cv, err = convert(fv, elemType, cfg, subPath)
		}
		if err != nil {
			return reflect.Value{}, err
		}
		out.SetMapIndex(reflect.ValueOf(key).Convert(dstType.Key()), cv)
	}
	return out, nil
}

// fieldValueByIndex 按嵌套索引读取字段；路径上遇到 nil 指针时返回不可用标记。
// 用于安全处理 nil 指针嵌入字段，避免 FieldByIndex 直接 panic。
func fieldValueByIndex(v reflect.Value, index []int) (reflect.Value, bool) {
	for _, idx := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}, false
			}
			v = v.Elem()
		}
		v = v.Field(idx)
	}
	return v, true
}

// fieldValueForSet 按嵌套索引定位可写字段；路径上的 nil 指针嵌入字段自动分配。
func fieldValueForSet(v reflect.Value, index []int) (reflect.Value, bool) {
	for _, idx := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				if !v.CanSet() {
					// 无法通过反射分配（例如未导出嵌入指针）时跳过该字段，避免 panic。
					return reflect.Value{}, false
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(idx)
	}
	if !v.CanSet() {
		return reflect.Value{}, false
	}
	return v, true
}
