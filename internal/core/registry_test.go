package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

// testStatus 用于注册表测试的自定义类型。
type testStatus int

// resetRegistryForTest 清空全局注册表，保证测试之间互不影响。
func resetRegistryForTest() {
	globalRegistry.mu.Lock()
	globalRegistry.table = make(map[convKey]converterFunc)
	globalRegistry.mu.Unlock()
}

func TestRegisterConverter(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	// 注册自定义转换器：Status → string。
	err := RegisterConverter(func(s testStatus) (string, error) {
		if s == 0 {
			return "", errors.New("未知状态")
		}
		return fmt.Sprintf("status-%d", s), nil
	})
	testx.NoError(t, err)

	got, err := To[string](testStatus(2))
	testx.NoError(t, err)
	testx.Equal(t, got, "status-2")

	// string → Status 注册（使用不同的键）。
	err = RegisterConverter(func(s string) (testStatus, error) {
		switch s {
		case "active":
			return testStatus(1), nil
		default:
			return 0, errors.New("未知状态名")
		}
	})
	testx.NoError(t, err)

	st, err := To[testStatus]("active")
	testx.NoError(t, err)
	testx.Equal(t, st, testStatus(1))
}

func TestRegisterDuplicate(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	err := RegisterConverter(func(s testStatus) (int, error) {
		return int(s), nil
	})
	testx.NoError(t, err)

	// 重复注册同一 (源, 目标) 返回 ErrDuplicateConverter。
	err = RegisterConverter(func(s testStatus) (int, error) {
		return 0, nil
	})
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrDuplicateConverter)
}

func TestRegisterConverterPriority(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	// 注册 int → string 覆盖内置转换。
	err := RegisterConverter(func(s int) (string, error) {
		return fmt.Sprintf("custom-%d", s), nil
	})
	testx.NoError(t, err)

	got, err := To[string](42)
	testx.NoError(t, err)
	testx.Equal(t, got, "custom-42")
}

func TestRegisterConverterErrorPassThrough(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	sourceErr := errors.New("内部错误")
	err := RegisterConverter(func(s testStatus) (bool, error) {
		return false, sourceErr
	})
	testx.NoError(t, err)

	_, err = To[bool](testStatus(1))
	testx.Error(t, err)
	var e *errx.Error
	testx.True(t, errors.As(err, &e))
	// 原始错误链保留，可解包到源错误。
	testx.True(t, errors.Is(err, sourceErr))
}

func TestRegisterConverterPanicRecovered(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	err := RegisterConverter(func(s testStatus) (float64, error) {
		panic("转换器崩溃")
	})
	testx.NoError(t, err)

	_, err = To[float64](testStatus(3))
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestRegisterConverterInterfaceTypeRejected(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	err := RegisterConverter(func(s any) (string, error) {
		return "x", nil
	})
	testx.Error(t, err)
	testx.ErrCode(t, err, ErrConversionFailed)
}

func TestRegisterConverterConcurrentRead(t *testing.T) {
	resetRegistryForTest()
	defer resetRegistryForTest()

	err := RegisterConverter(func(s testStatus) (string, error) {
		return "ok", nil
	})
	testx.NoError(t, err)

	// 并发查找不产生数据竞争。
	testx.Concurrently(t, 32, func() {
		got, err := To[string](testStatus(1))
		testx.NoError(t, err)
		testx.Equal(t, got, "ok")
	})
}
