package core

import (
	"fmt"
	"reflect"
)

// Convert 是核心调度入口：将 src 转换为 dstType 类型的值。
// 供根包公开 API 与内部递归调用。
// dstType 为 nil（目标为 any）时直接返回源值；配置默认值时零值源返回默认值。
func Convert(src any, dstType reflect.Type, opts ...Option) (any, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return convertTop(src, dstType, &cfg)
}

// convertTop 顶层转换：nil 检查 → 调度 → 校验。
func convertTop(src any, dstType reflect.Type, cfg *convertConfig) (any, error) {
	// 目标为接口（any）：无需转换，直接返回源值。
	if dstType == nil {
		if cfg.defaultSet && (src == nil || isZeroValue(src)) {
			return cfg.defaultValue, nil
		}
		return src, nil
	}
	// 源为 nil。
	if src == nil {
		if cfg.defaultSet {
			return convertDefaultValue(cfg, dstType)
		}
		if dstType.Kind() == reflect.Ptr {
			return reflect.Zero(dstType).Interface(), nil
		}
		return reflect.Zero(dstType).Interface(),
			newError(ErrConversionFailed, "源值为 nil", nil, dstType, "", src)
	}

	srcVal := reflect.ValueOf(src)

	// 源链中包含 nil 指针/接口：默认值、错误模式或零值。
	if hasNilPointer(srcVal) {
		if cfg.defaultSet {
			return convertDefaultValue(cfg, dstType)
		}
		if cfg.nilPointerAsError {
			return reflect.Zero(dstType).Interface(),
				newError(ErrNilPointer, "源值为 nil 指针或接口", srcVal.Type(), dstType, "", src)
		}
		if dstType.Kind() == reflect.Ptr {
			return reflect.Zero(dstType).Interface(), nil
		}
		return reflect.Zero(dstType).Interface(), nil
	}

	out, err := convert(srcVal, dstType, cfg, "")
	if err != nil {
		if cfg.defaultSet {
			return convertDefaultValue(cfg, dstType)
		}
		return nil, err
	}

	// 结果为零值时按 WithDefault 语义返回默认值。
	if cfg.defaultSet && isZeroValue(out.Interface()) {
		return convertDefaultValue(cfg, dstType)
	}

	// 转换后校验。
	if cfg.validationSet {
		if verr := validateValue(out.Interface(), cfg.validationRules); verr != nil {
			if cfg.defaultSet {
				return convertDefaultValue(cfg, dstType)
			}
			return nil, verr
		}
	}
	return out.Interface(), nil
}

// convertDefaultValue 将 WithDefault 设置的默认值转换为目标类型。
// 默认值本身不参与默认值递归，也不参与结果校验。
func convertDefaultValue(cfg *convertConfig, dstType reflect.Type) (any, error) {
	if !cfg.defaultSet {
		return nil, nil
	}
	if dstType == nil {
		return cfg.defaultValue, nil
	}
	if cfg.defaultValue == nil {
		return reflect.Zero(dstType).Interface(), nil
	}

	dv := reflect.ValueOf(cfg.defaultValue)
	if dv.Type() == dstType {
		return cfg.defaultValue, nil
	}

	cfg2 := *cfg
	cfg2.defaultSet = false
	cfg2.validationSet = false
	cfg2.path = nil
	out, err := convertTop(dv.Interface(), dstType, &cfg2)
	if err != nil {
		return nil, newError(ErrConversionFailed,
			fmt.Sprintf("默认值类型不匹配：%s 无法转换为 %s", dv.Type(), dstType),
			dv.Type(), dstType, "", cfg.defaultValue)
	}
	return out, nil
}

// convert 将反射值 srcVal 转换为 dstType 类型（内部递归入口）。
// 处理指针/接口解引用、目标指针重建、自定义转换器与内置转换分发。
func convert(srcVal reflect.Value, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	// 记录本次解引用的指针地址，返回前从路径集合移除（兄弟字段共享指针不误报）。
	var addedPtrs []uintptr
	defer func() {
		for _, p := range addedPtrs {
			delete(cfg.path, p)
		}
	}()

	// 保存原始目标类型：nil 源需按原始类型取零值（指针目标为 nil 指针，
	// 与顶层 convertTop 的 nil 语义保持一致）。
	dstOrig := dstType

	// 计算目标指针层数，解出元素类型。
	dstPtrDepth := 0
	for dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
		dstPtrDepth++
	}

	// 解引用源指针与接口，处理 nil，检测循环引用。
	for srcVal.Kind() == reflect.Ptr || srcVal.Kind() == reflect.Interface {
		if srcVal.IsNil() {
			if cfg.nilPointerAsError {
				return reflect.Value{}, newError(ErrNilPointer, "源值为 nil 指针或接口",
					srcVal.Type(), dstType, fieldPath, nil)
			}
			// 默认模式：nil 指针/接口转为目标原始类型的零值。
			// 若目标被剥掉过指针层，必须按原始类型取零值，
			// 否则返回元素零值会导致调用方 Set 时类型不匹配 panic。
			return reflect.Zero(dstOrig), nil
		}
		// 目标为接口时，优先保留当前已实现该接口的源类型（含指针方法集）。
		if dstType.Kind() == reflect.Interface && srcVal.Type().Implements(dstType) {
			return interfaceResult(srcVal, dstType, dstPtrDepth, fieldPath)
		}
		if srcVal.Kind() == reflect.Ptr {
			ptr := srcVal.Pointer()
			if cfg.path == nil {
				cfg.path = make(map[uintptr]bool)
			}
			if cfg.path[ptr] {
				return reflect.Value{}, newError(ErrCircularReference, "检测到循环引用",
					srcVal.Type(), dstType, fieldPath, nil)
			}
			cfg.path[ptr] = true
			addedPtrs = append(addedPtrs, ptr)
		}
		srcVal = srcVal.Elem()
	}
	srcType := srcVal.Type()

	// 非指针/接口源值直接命中接口目标时，同样需要实现检查。
	if dstType.Kind() == reflect.Interface {
		return interfaceResult(srcVal, dstType, dstPtrDepth, fieldPath)
	}

	// 元素级类型匹配：直接返回（值复制）。
	if srcType == dstType {
		return buildPointer(srcVal, dstPtrDepth), nil
	}

	// 自定义转换器（元素级）。
	if fn, ok := lookupConverter(srcType, dstType); ok {
		out, err := fn(srcVal.Interface())
		if err != nil {
			// 转换器错误通过 errx 封装，保留原始错误链。
			return reflect.Value{}, wrapError(err, ErrConversionFailed,
				srcType, dstType, fieldPath, srcVal.Interface())
		}
		return buildPointer(reflect.ValueOf(out), dstPtrDepth), nil
	}

	// 内置转换分发（元素级）。
	out, err := convertBuiltin(srcVal, srcType, dstType, cfg, fieldPath)
	if err != nil {
		return reflect.Value{}, err
	}
	return buildPointer(out, dstPtrDepth), nil
}

// convertBuiltin 按目标类型 Kind 分发到内置转换器。
func convertBuiltin(srcVal reflect.Value, srcType, dstType reflect.Type, cfg *convertConfig, fieldPath string) (reflect.Value, error) {
	// 源为时间类型（time.Time / time.Duration）的特殊转换。
	if srcType == timeType || srcType == durationType {
		return convertFromTime(srcVal, srcType, dstType, cfg, fieldPath)
	}
	// 目标为时间类型（time.Duration 的 Kind 是 Int64，需在标量分发之前处理）。
	if dstType == timeType {
		return convertToTimeValue(srcVal, srcType, cfg, fieldPath)
	}
	if dstType == durationType {
		return convertToDurationValue(srcVal, srcType, cfg, fieldPath)
	}

	switch dstType.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return convertScalar(srcVal, srcType, dstType, cfg, fieldPath)
	case reflect.Struct:
		return convertStruct(srcVal, dstType, cfg, fieldPath)
	case reflect.Map:
		return convertMap(srcVal, dstType, cfg, fieldPath)
	case reflect.Slice, reflect.Array:
		return convertSlice(srcVal, dstType, cfg, fieldPath)
	case reflect.Interface:
		if srcType.Implements(dstType) {
			return srcVal, nil
		}
		return reflect.Value{}, newError(ErrConversionFailed,
			fmt.Sprintf("源类型 %s 未实现接口 %s", srcType, dstType),
			srcType, dstType, fieldPath, srcVal.Interface())
	}
	return reflect.Value{}, newError(ErrConversionFailed,
		fmt.Sprintf("不支持的目标类型：%s", dstType), srcType, dstType, fieldPath, srcVal.Interface())
}

// interfaceResult 将实现了目标接口的源值包装为接口结果。
// 目标为普通接口时返回源值本身；目标为指针接口时构造对应层级的接口指针。
func interfaceResult(srcVal reflect.Value, dstType reflect.Type, dstPtrDepth int, fieldPath string) (reflect.Value, error) {
	if !srcVal.Type().Implements(dstType) {
		return reflect.Value{}, newError(ErrConversionFailed,
			fmt.Sprintf("源类型 %s 未实现接口 %s", srcVal.Type(), dstType),
			srcVal.Type(), dstType, fieldPath, srcVal.Interface())
	}
	if dstPtrDepth == 0 {
		return srcVal, nil
	}
	iv := reflect.New(dstType).Elem()
	iv.Set(srcVal)
	return buildPointer(iv, dstPtrDepth), nil
}
