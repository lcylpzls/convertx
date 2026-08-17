package convertx

import (
	"time"

	"github.com/lcylpzls/convertx/internal/core"
)

// WithDefault 设置在转换失败、结果为零值或源为 nil 时返回的默认值。
// 使用该选项后，转换函数在错误场景返回默认值且不再返回错误。
func WithDefault(v any) Option { return core.WithDefault(v) }

// WithStrictMode 开启严格模式：结构体/map 转换遇到未匹配字段时返回错误。
func WithStrictMode() Option { return core.WithStrictMode() }

// WithTagName 指定结构体转换使用的 tag 名，默认 "convertx"。
func WithTagName(name string) Option { return core.WithTagName(name) }

// WithValidation 设置在转换成功后对结果执行的 validx 校验规则。
func WithValidation(rules string) Option { return core.WithValidation(rules) }

// WithTimeLayout 指定时间解析格式，设置后跳过自动格式识别。
func WithTimeLayout(layout string) Option { return core.WithTimeLayout(layout) }

// WithTimeUnit 指定时间戳精度，设置后跳过自动精度推断。
func WithTimeUnit(unit TimeUnit) Option { return core.WithTimeUnit(unit) }

// WithDefaultLocation 设置无时区时间字符串的默认时区，默认 time.Local。
func WithDefaultLocation(loc *time.Location) Option {
	return core.WithDefaultLocation(loc)
}

// WithNilPointerAsError 开启 nil 指针报错模式：遇到 nil 指针返回错误而非零值。
func WithNilPointerAsError() Option { return core.WithNilPointerAsError() }
