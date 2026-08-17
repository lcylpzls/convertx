package core

import (
	"fmt"
	"reflect"

	"github.com/lcylpzls/errx"
)

// 错误上下文字段的 key。
const (
	fieldKeySourceType   = "source_type"
	fieldKeyTargetType   = "target_type"
	fieldKeyFieldPath    = "field_path"
	fieldKeyValuePreview = "value_preview"
)

// valuePreviewLimit 是 value_preview 的最大长度，避免泄露完整敏感值。
const valuePreviewLimit = 100

// wrapError 将内部错误封装为携带转换上下文的 errx.Error。
// code 为空时使用 ErrConversionFailed；err 为 nil 时返回 nil。
func wrapError(err error, code errx.Code, srcType, dstType reflect.Type, fieldPath string, src any) error {
	if err == nil {
		return nil
	}
	if code == "" {
		code = ErrConversionFailed
	}
	e := errx.WrapCode(err, code, err.Error())
	return attachContext(e, srcType, dstType, fieldPath, src)
}

// newError 构造无底层错误来源的封装错误。
// code 为空时使用 ErrConversionFailed。
func newError(code errx.Code, msg string, srcType, dstType reflect.Type, fieldPath string, src any) error {
	if code == "" {
		code = ErrConversionFailed
	}
	e := errx.NewCode(code, msg)
	return attachContext(e, srcType, dstType, fieldPath, src)
}

// attachContext 为 errx.Error 附加转换上下文字段。
// field_path 始终携带（顶层转换为空字符串），保证错误结构一致。
func attachContext(e *errx.Error, srcType, dstType reflect.Type, fieldPath string, src any) *errx.Error {
	if srcType != nil {
		e = e.WithField(fieldKeySourceType, srcType.String())
	}
	if dstType != nil {
		e = e.WithField(fieldKeyTargetType, dstType.String())
	}
	e = e.WithField(fieldKeyFieldPath, fieldPath)
	e = e.WithField(fieldKeyValuePreview, preview(src))
	return e
}

// preview 生成源值摘要，超长时截断至 valuePreviewLimit 字符。
func preview(v any) string {
	s := fmt.Sprintf("%v", v)
	if len(s) > valuePreviewLimit {
		return s[:valuePreviewLimit]
	}
	return s
}

// withFieldPath 将字段路径拼接到已有错误上（浅层优先，保持嵌套完整）。
func withFieldPath(err error, fieldPath string) error {
	if err == nil || fieldPath == "" {
		return err
	}
	if e, ok := errx.As(err); ok && e != nil {
		return e.WithField(fieldKeyFieldPath, fieldPath)
	}
	return err
}
