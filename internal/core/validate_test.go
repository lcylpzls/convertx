package core

import (
	"testing"

	"github.com/lcylpzls/testx"
)

func TestToAndValidatePass(t *testing.T) {
	v, err := ToAndValidate[int]("42", "min=0,max=100")
	testx.NoError(t, err)
	testx.Equal(t, v, 42)
}

func TestToAndValidateFail(t *testing.T) {
	_, err := ToAndValidate[int]("150", "min=0,max=100")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrValidationFailed)
}

func TestToAndValidateConversionFail(t *testing.T) {
	_, err := ToAndValidate[int]("abc", "min=0")
	testx.Error(t, err)
	// 转换失败返回转换错误，不执行校验。
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestWithValidationOption(t *testing.T) {
	// 选项模式。
	v, err := To[int]("42", WithValidation("min=0,max=100"))
	testx.NoError(t, err)
	testx.Equal(t, v, 42)

	_, err = To[int]("200", WithValidation("min=0,max=100"))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrValidationFailed)
}

func TestValidate(t *testing.T) {
	testx.NoError(t, Validate(42, "min=0,max=100"))
	testx.NoError(t, Validate("user@example.com", "email"))

	err := Validate("not-email", "email")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrValidationFailed)
}

func TestValidateField(t *testing.T) {
	testx.NoError(t, ValidateField("abc", "min=2"))

	err := ValidateField("a", "min=2")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrValidationFailed)
}

func TestValidateInvalidRule(t *testing.T) {
	err := Validate(42, "invalid_rule_name=1")
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrValidationFailed)
}
