package core

import "github.com/lcylpzls/errx"

// 错误码定义：convertx 各失败场景的错误码。
// 统一使用 CONVERTX_ 前缀，避免与家族其他库的全局错误码冲突。
const (
	// ErrConversionFailed 通用转换失败：类型不兼容、格式非法等。
	ErrConversionFailed errx.Code = "CONVERTX_CONVERSION_FAILED"
	// ErrValidationFailed 校验失败：validx 校验规则不通过。
	ErrValidationFailed errx.Code = "CONVERTX_VALIDATION_FAILED"
	// ErrOverflow 数值溢出：源值超出目标类型范围。
	ErrOverflow errx.Code = "CONVERTX_OVERFLOW"
	// ErrUnknownFormat 时间格式无法识别。
	ErrUnknownFormat errx.Code = "CONVERTX_UNKNOWN_FORMAT"
	// ErrCircularReference 结构体转换中检测到循环引用。
	ErrCircularReference errx.Code = "CONVERTX_CIRCULAR_REFERENCE"
	// ErrDuplicateConverter 同一 (源类型, 目标类型) 转换器重复注册。
	ErrDuplicateConverter errx.Code = "CONVERTX_DUPLICATE_CONVERTER"
	// ErrNilPointer 开启 nil 指针报错模式时遇到 nil 指针。
	ErrNilPointer errx.Code = "CONVERTX_NIL_POINTER"
	// ErrUnmatchedField 严格模式下结构体/map 存在未匹配字段。
	ErrUnmatchedField errx.Code = "CONVERTX_UNMATCHED_FIELD"
)

// init 注册错误码及其 Kind 分类，供 errx.NewCode / WrapCode 使用。
func init() {
	registerCode(ErrConversionFailed, "类型转换失败", errx.KindInvalid)
	registerCode(ErrValidationFailed, "校验失败", errx.KindInvalid)
	registerCode(ErrOverflow, "数值溢出", errx.KindInvalid)
	registerCode(ErrUnknownFormat, "时间格式无法识别", errx.KindInvalid)
	registerCode(ErrCircularReference, "结构体循环引用", errx.KindInvalid)
	registerCode(ErrDuplicateConverter, "转换器重复注册", errx.KindAlreadyExists)
	registerCode(ErrNilPointer, "空指针", errx.KindInvalid)
	registerCode(ErrUnmatchedField, "未匹配字段", errx.KindInvalid)
}

// registerCode 注册错误码说明与 Kind 分类。
func registerCode(code errx.Code, desc string, kind errx.Kind) {
	errx.RegisterCode(code, desc)
	errx.RegisterCodeKind(code, kind)
}
