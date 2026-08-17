package core

import (
	"github.com/lcylpzls/validx"
)

// validateValue 委托 validx 校验单个值，并将校验错误重新封装为
// convertx 的错误码（ErrValidationFailed），保留原始 validx 错误链。
func validateValue(value any, rules string) error {
	if err := validx.ValidateField(value, rules); err != nil {
		return wrapError(err, ErrValidationFailed, nil, nil, "", value)
	}
	return nil
}
