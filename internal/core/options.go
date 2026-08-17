package core

import "time"

// TimeUnit 表示时间戳精度。
type TimeUnit int

// 时间戳精度枚举。
const (
	// UnitSecond 秒。
	UnitSecond TimeUnit = iota
	// UnitMillisecond 毫秒。
	UnitMillisecond
	// UnitMicrosecond 微秒。
	UnitMicrosecond
	// UnitNanosecond 纳秒。
	UnitNanosecond
)

// convertConfig 是转换选项的聚合配置。
type convertConfig struct {
	defaultValue       any
	defaultSet         bool
	strictMode         bool
	tagName            string
	validationRules    string
	validationSet      bool
	timeLayout         string
	timeUnit           TimeUnit
	timeUnitSet        bool
	defaultLocation    *time.Location
	defaultLocationSet bool
	nilPointerAsError  bool
	// path 是当前递归路径上已访问的指针地址集合，用于循环引用检测。
	path map[uintptr]bool
}

// defaultConfig 返回默认转换配置。
func defaultConfig() convertConfig {
	return convertConfig{tagName: "convertx"}
}

// Option 是转换选项函数，可组合传入各转换 API。
type Option func(*convertConfig)

// WithDefault 设置在转换失败、结果为零值或源为 nil 时返回的默认值。
// 使用该选项后，转换函数在错误场景返回默认值且不再返回错误。
func WithDefault(v any) Option {
	return func(c *convertConfig) {
		c.defaultValue = v
		c.defaultSet = true
	}
}

// WithStrictMode 开启严格模式：结构体/map 转换遇到未匹配字段时返回错误。
func WithStrictMode() Option {
	return func(c *convertConfig) {
		c.strictMode = true
	}
}

// WithTagName 指定结构体转换使用的 tag 名，默认 "convertx"。
func WithTagName(name string) Option {
	return func(c *convertConfig) {
		c.tagName = name
	}
}

// WithValidation 设置在转换成功后对结果执行的 validx 校验规则。
func WithValidation(rules string) Option {
	return func(c *convertConfig) {
		c.validationRules = rules
		c.validationSet = true
	}
}

// WithTimeLayout 指定时间解析格式，设置后跳过自动格式识别。
func WithTimeLayout(layout string) Option {
	return func(c *convertConfig) {
		c.timeLayout = layout
	}
}

// WithTimeUnit 指定时间戳精度，设置后跳过自动精度推断。
func WithTimeUnit(unit TimeUnit) Option {
	return func(c *convertConfig) {
		c.timeUnit = unit
		c.timeUnitSet = true
	}
}

// WithDefaultLocation 设置无时区时间字符串的默认时区，默认 time.Local。
func WithDefaultLocation(loc *time.Location) Option {
	return func(c *convertConfig) {
		c.defaultLocation = loc
		c.defaultLocationSet = true
	}
}

// WithNilPointerAsError 开启 nil 指针报错模式：遇到 nil 指针返回错误而非零值。
func WithNilPointerAsError() Option {
	return func(c *convertConfig) {
		c.nilPointerAsError = true
	}
}
