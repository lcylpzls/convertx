// basic 示例：convertx 基础类型转换、结构体转换、校验一体化、
// 自定义转换器注册与时间时区处理。
package main

import (
	"fmt"
	"time"

	"github.com/lcylpzls/convertx"
)

// UserDO 是领域模型（数据对象）。
type UserDO struct {
	ID     int64
	Name   string
	Age    int
	Active bool
}

// UserVO 是视图对象，字段类型与 UserDO 不同。
type UserVO struct {
	ID     string
	Name   string
	Age    string
	Active string
}

// Status 是业务自定义类型，演示自定义转换器。
type Status int

// String 返回状态的字符串表示。
func (s Status) String() string {
	switch s {
	case 1:
		return "active"
	case 0:
		return "inactive"
	default:
		return "unknown"
	}
}

func init() {
	// 注册自定义转换器：Status → string，优先级高于内置。
	err := convertx.RegisterConverter(func(s Status) (string, error) {
		return s.String(), nil
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	basicConversion()
	structConversion()
	validation()
	customConverter()
	timeConversion()
}

// basicConversion 基础类型转换与默认值。
func basicConversion() {
	// 泛型转换入口。
	v, err := convertx.ToInt("42")
	if err != nil {
		panic(err)
	}
	fmt.Printf("ToInt(\"42\") = %d\n", v)

	// 溢出检测。
	if _, err := convertx.ToInt8(128); err != nil {
		fmt.Printf("ToInt8(128) 错误码: %v\n", err)
	}

	// 带默认值（转换失败或零值时返回默认值）。
	fmt.Printf("ToOrDefault(\"abc\", -1) = %d\n", convertx.ToOrDefault[int]("abc", -1))
}

// structConversion 结构体转换（含自动类型转换与 map 互转）。
func structConversion() {
	do := UserDO{ID: 1, Name: "Alice", Age: 30, Active: true}
	var vo UserVO
	if err := convertx.Struct(&do, &vo); err != nil {
		panic(err)
	}
	fmt.Printf("Struct: %+v\n", vo)

	// struct → map。
	m, err := convertx.StructToMap(do)
	if err != nil {
		panic(err)
	}
	fmt.Printf("StructToMap: %v\n", m)

	// map → struct。
	var back UserDO
	if err := convertx.MapToStruct(map[string]any{"ID": "2", "Name": "Bob"}, &back); err != nil {
		panic(err)
	}
	fmt.Printf("MapToStruct: %+v\n", back)
}

// validation 转换后自动校验与独立校验。
func validation() {
	// 转换 + 校验一体化。
	v, err := convertx.ToAndValidate[int]("42", "min=0,max=100")
	if err != nil {
		panic(err)
	}
	fmt.Printf("ToAndValidate(\"42\") = %d\n", v)

	// 校验失败。
	if err := convertx.Validate("not-email", "email"); err != nil {
		fmt.Printf("Validate 错误码: %v\n", err)
	}
}

// customConverter 自定义转换器。
func customConverter() {
	s, err := convertx.ToString(Status(1))
	if err != nil {
		panic(err)
	}
	fmt.Printf("自定义转换 Status(1) → %q\n", s)
}

// timeConversion 时间解析、时区与时间戳。
func timeConversion() {
	// 自动识别 RFC3339 格式。
	t, err := convertx.ToTime("2026-08-17T10:00:00+08:00")
	if err != nil {
		panic(err)
	}
	fmt.Printf("ToTime: %s\n", t.Format("2006-01-02 15:04:05 MST"))

	// 时间戳精度自动推断（秒级）。
	ts, err := convertx.ToInt64(t)
	if err != nil {
		panic(err)
	}
	fmt.Printf("时间戳(秒): %d\n", ts)

	// 时区转换。
	fmt.Printf("UTC: %s\n", convertx.ToUTC(t).Format(time.RFC3339))
}
