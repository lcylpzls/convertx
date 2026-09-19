# API 接口文档

| 项目 | convertx |
|------|----------|
| 文档版本 | v1.0 |
| 编制日期 | 2026-08-17 |
| 前置文档 | 03-prd.md、05-architecture-design.md |
| 模块路径 | `github.com/lcylpzls/convertx` |

---

## 1. 文档说明

本文档定义 convertx 库的全部公开 API，是开发调用和测试的唯一依据。所有函数遵循以下统一约定：

- **错误返回**：所有可能失败的函数均返回 `(T, error)` 或 `error`，无仅返值变体
- **错误类型**：所有 `error` 均为 `*errx.Error`，可通过 `errors.As` 提取
- **禁止 panic**：任何输入（含 nil、非法值）均不 panic
- **泛型优先**：优先提供泛型 API `To[T]()`，同时提供类型特定函数便于老代码迁移

---

## 2. 常量与类型

### 2.1 版本常量

```go
const Version = "1.0.0"
```

### 2.2 错误码常量

| 常量名 | 值 | 含义 |
|--------|-----|------|
| `ErrConversionFailed` | "CONVERSION_FAILED" | 通用转换失败 |
| `ErrValidationFailed` | "VALIDATION_FAILED" | 校验失败 |
| `ErrOverflow` | "OVERFLOW" | 数值溢出 |
| `ErrUnknownFormat` | "UNKNOWN_FORMAT" | 时间格式无法识别 |
| `ErrCircularReference` | "CIRCULAR_REFERENCE" | 结构体循环引用 |
| `ErrDuplicateConverter` | "DUPLICATE_CONVERTER" | 转换器重复注册 |
| `ErrNilPointer` | "NIL_POINTER" | 空指针（开启报错模式时） |
| `ErrUnmatchedField` | "UNMATCHED_FIELD" | 严格模式下未匹配字段 |

### 2.3 核心类型

| 类型 | 定义 | 说明 |
|------|------|------|
| `Option` | `func(*convertConfig)` | 转换选项函数类型 |
| `TimeUnit` | `int` 枚举 | 时间戳精度：`UnitSecond` / `UnitMillisecond` / `UnitMicrosecond` / `UnitNanosecond` |

---

## 3. 基础类型转换

### 3.1 泛型转换入口

#### `func To[T any](src any, opts ...Option) (T, error)`

将任意源值转换为目标类型 `T`。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `any` | 是 | 源值，支持任意基础类型、指针、时间类型 |
| opts | `...Option` | 否 | 转换选项（见第9节） |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `T` | 转换后的值，失败时为 T 的零值 |
| 错误 | `error` | 失败时为 `*errx.Error`，携带 source_type/target_type |

**错误码**：`ErrConversionFailed`、`ErrOverflow`、`ErrUnknownFormat`

**示例**：
```go
v, err := convertx.To[int]("42")          // v=42, err=nil
v, err := convertx.To[string](123)        // v="123", err=nil
v, err := convertx.To[bool]("true")       // v=true, err=nil
v, err := convertx.To[int8](300)          // v=0, err=ErrOverflow
```

---

### 3.2 类型特定转换函数

以下函数为 `To[T]` 的便捷别名，行为完全一致：

| 函数签名 | 说明 |
|---------|------|
| `func ToInt(src any, opts ...Option) (int, error)` | 转 int |
| `func ToInt8(src any, opts ...Option) (int8, error)` | 转 int8 |
| `func ToInt16(src any, opts ...Option) (int16, error)` | 转 int16 |
| `func ToInt32(src any, opts ...Option) (int32, error)` | 转 int32 |
| `func ToInt64(src any, opts ...Option) (int64, error)` | 转 int64 |
| `func ToUint(src any, opts ...Option) (uint, error)` | 转 uint |
| `func ToUint8(src any, opts ...Option) (uint8, error)` | 转 uint8 |
| `func ToUint16(src any, opts ...Option) (uint16, error)` | 转 uint16 |
| `func ToUint32(src any, opts ...Option) (uint32, error)` | 转 uint32 |
| `func ToUint64(src any, opts ...Option) (uint64, error)` | 转 uint64 |
| `func ToFloat32(src any, opts ...Option) (float32, error)` | 转 float32 |
| `func ToFloat64(src any, opts ...Option) (float64, error)` | 转 float64 |
| `func ToBool(src any, opts ...Option) (bool, error)` | 转 bool |
| `func ToString(src any, opts ...Option) (string, error)` | 转 string |
| `func ToByte(src any, opts ...Option) (byte, error)` | 转 byte（=uint8） |
| `func ToRune(src any, opts ...Option) (rune, error)` | 转 rune（=int32） |

---

### 3.3 带默认值转换

#### `func ToOrDefault[T any](src any, def T) T`

转换失败或源为零值/nil 时返回默认值，永不返回 error。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `any` | 是 | 源值 |
| def | `T` | 是 | 默认值，必须与目标类型一致 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `T` | 转换成功返转换值；失败/零值返默认值 |

**触发默认值的条件**：源为 nil、源为类型零值（`""`/`0`/`false`）、转换过程出错。

**示例**：
```go
v := convertx.ToOrDefault[int]("42", -1)    // v=42
v := convertx.ToOrDefault[int]("abc", -1)   // v=-1（转换失败）
v := convertx.ToOrDefault[int]("0", -1)     // v=-1（零值触发）
v := convertx.ToOrDefault[string](nil, "N/A") // v="N/A"
```

---

## 4. 结构体转换

### 4.1 struct → struct

#### `func Struct(src any, dst any, opts ...Option) error`

将源结构体的字段复制到目标结构体，按字段名/tag 匹配，类型不同时自动转换。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `any` | 是 | 源结构体值或指针 |
| dst | `any` | 是 | 目标结构体指针，必须非 nil |
| opts | `...Option` | 否 | 选项（`WithTagName`、`WithStrictMode` 等） |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 错误 | `error` | 失败时为 `*errx.Error`，携带 field_path |

**错误码**：`ErrConversionFailed`、`ErrCircularReference`、`ErrUnmatchedField`、`ErrOverflow`

**示例**：
```go
type Source struct {
    Name string
    Age  int
}
type Target struct {
    Name string
    Age  string  // 类型不同，自动转换
}

s := Source{Name: "Alice", Age: 30}
var t Target
err := convertx.Struct(&s, &t)
// t = {Name:"Alice", Age:"30"}
```

#### `func ToStruct[T any](src any, opts ...Option) (T, error)`

泛型版本，返回目标结构体值。

```go
t, err := convertx.ToStruct[Target](&s)
```

---

### 4.2 struct → map

#### `func StructToMap(src any, opts ...Option) (map[string]any, error)`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `any` | 是 | 源结构体值或指针 |
| opts | `...Option` | 否 | 选项（`WithTagName` 控制 map key 来源） |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| map | `map[string]any` | 转换后的 map |
| 错误 | `error` | 失败时为 `*errx.Error` |

**示例**：
```go
type User struct {
    Name string `convertx:"username"`
    Age  int
}
m, err := convertx.StructToMap(User{Name: "Bob", Age: 25})
// m = map[string]any{"username": "Bob", "Age": 25}
```

---

### 4.3 map → struct

#### `func MapToStruct(src map[string]any, dst any, opts ...Option) error`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `map[string]any` | 是 | 源 map，nil map 视为空 map |
| dst | `any` | 是 | 目标结构体指针 |
| opts | `...Option` | 否 | 选项 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 错误 | `error` | 失败时为 `*errx.Error` |

**支持强类型 map**：`map[string]string`、`map[string]int` 等也可作为 src 传入（通过 `any` 参数）。

**示例**：
```go
m := map[string]any{"Name": "Carol", "Age": "28"}
var u User
err := convertx.MapToStruct(m, &u)
// u = {Name:"Carol", Age:28}
```

---

## 5. 集合类型转换

### 5.1 slice → slice

#### `func SliceToSlice[Src any, Dst any](src []Src, opts ...Option) ([]Dst, error)`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `[]Src` | 是 | 源 slice，nil 返回 nil |
| opts | `...Option` | 否 | 选项 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `[]Dst` | 转换后的 slice，保留原顺序 |
| 错误 | `error` | 任一元素转换失败则整体失败 |

**示例**：
```go
src := []string{"1", "2", "3"}
dst, err := convertx.SliceToSlice[string, int](src)
// dst = []int{1, 2, 3}
```

---

### 5.2 map → map

#### `func MapToMap[SK comparable, SV any, DK comparable, DV any](src map[SK]SV, opts ...Option) (map[DK]DV, error)`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `map[SK]SV` | 是 | 源 map，nil 返回 nil |
| opts | `...Option` | 否 | 选项 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `map[DK]DV` | 转换后的 map |
| 错误 | `error` | 任一 key/value 转换失败则整体失败 |

**示例**：
```go
src := map[string]string{"a": "1", "b": "2"}
dst, err := convertx.MapToMap[string, string, string, int](src)
// dst = map[string]int{"a":1, "b":2}
```

---

### 5.3 slice → map

#### `func SliceToMap[T any, K comparable](src []T, keyFn func(T) K, opts ...Option) (map[K]T, error)`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `[]T` | 是 | 源 slice |
| keyFn | `func(T) K` | 是 | 从元素提取 map key 的函数 |
| opts | `...Option` | 否 | 选项 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `map[K]T` | key 重复时后者覆盖前者 |
| 错误 | `error` | keyFn panic 时 recover 为错误 |

**示例**：
```go
type Item struct{ ID int; Name string }
items := []Item{{1,"a"},{2,"b"}}
m, err := convertx.SliceToMap(items, func(item Item) int { return item.ID })
// m = map[int]Item{1:{1,"a"}, 2:{2,"b"}}
```

---

### 5.4 map → slice

#### `func MapToSlice[K comparable, V any](src map[K]V, opts ...Option) ([]V, error)`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `map[K]V` | 是 | 源 map |
| opts | `...Option` | 否 | 选项 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `[]V` | 值的 slice，**顺序不保证**（Go map 迭代无序） |
| 错误 | `error` | 通常为 nil |

---

## 6. 时间类型转换

### 6.1 通用时间转换

#### `func ToTime(src any, opts ...Option) (time.Time, error)`

将 string / int64 / time.Time 转为 time.Time。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `any` | 是 | string（时间字符串）、int64（时间戳）、time.Time |
| opts | `...Option` | 否 | `WithTimeLayout`、`WithTimeUnit`、`WithDefaultLocation` |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `time.Time` | 转换后的时间 |
| 错误 | `error` | `ErrUnknownFormat`（格式无法识别） |

**string 自动识别格式**（按优先级）：RFC3339/Nano → `2006-01-02 15:04:05` → `2006-01-02` → RFC822/Z

**int64 精度自动推断**：<1e11秒 / 1e11~1e14毫秒 / 1e14~1e17微秒 / ≥1e17纳秒

**示例**：
```go
t, err := convertx.ToTime("2026-08-17T10:00:00+08:00")  // RFC3339
t, err := convertx.ToTime("2026-08-17 10:00:00")          // 默认Local
t, err := convertx.ToTime(int64(1755484800))              // 秒级时间戳
t, err := convertx.ToTime("2026/08/17", convertx.WithTimeLayout("2006/01/02"))
```

---

### 6.2 时间格式化

#### `func TimeToString(t time.Time, layout string) string`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| t | `time.Time` | 是 | 源时间 |
| layout | `string` | 是 | Go 时间布局，如 `time.RFC3339`、`"2006-01-02 15:04:05"` |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `string` | 格式化后的字符串，永不失败 |

---

### 6.3 时间戳转换

#### `func TimestampToTime(ts int64, unit TimeUnit) time.Time`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ts | `int64` | 是 | 时间戳数值 |
| unit | `TimeUnit` | 是 | `UnitSecond` / `UnitMillisecond` / `UnitMicrosecond` / `UnitNanosecond` |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `time.Time` | UTC 时间，永不失败 |

#### `func TimeToTimestamp(t time.Time, unit TimeUnit) int64`

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `int64` | 指定精度的时间戳 |

---

### 6.4 时区转换

#### `func ToUTC(t time.Time) time.Time`
转换为 UTC 时区。

#### `func ToLocal(t time.Time) time.Time`
转换为本地时区。

#### `func ToLocation(t time.Time, loc *time.Location) time.Time`
转换为指定时区。

---

### 6.5 Duration 转换

#### `func ToDuration(src any, opts ...Option) (time.Duration, error)`

支持 string（如 `"1h30m"`、`"500ms"`）、int64（纳秒）、`time.Duration` 互转。

---

## 7. 校验相关 API

### 7.1 转换后自动校验

#### `func ToAndValidate[T any](src any, rules string, opts ...Option) (T, error)`

转换成功后自动对结果执行 validx 校验。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | `any` | 是 | 源值 |
| rules | `string` | 是 | validx 规则字符串，如 `"min=0,max=100"` |
| opts | `...Option` | 否 | 其他选项 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 结果 | `T` | 转换且校验通过的值 |
| 错误 | `error` | 转换失败返转换错误；校验失败返 `ErrValidationFailed` |

**执行顺序**：转换 → 成功才校验 → 校验通过才返回值。

**示例**：
```go
v, err := convertx.ToAndValidate[int]("42", "min=0,max=100")  // v=42, err=nil
v, err := convertx.ToAndValidate[int]("150", "min=0,max=100") // err=ErrValidationFailed
v, err := convertx.ToAndValidate[int]("abc", "min=0")         // err=ErrConversionFailed
```

---

### 7.2 独立校验

#### `func Validate(value any, rules string) error`

对已有值执行校验，不做转换。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| value | `any` | 是 | 待校验值 |
| rules | `string` | 是 | validx 规则字符串 |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 错误 | `error` | 校验通过返 nil；失败返 `ErrValidationFailed` |

#### `func ValidateField(value any, rules string) error`

单字段校验，行为同 `Validate`，错误信息标记为字段级。

**示例**：
```go
err := convertx.Validate(42, "min=0,max=100")           // nil
err := convertx.Validate("not-email", "required,email") // ErrValidationFailed
```

---

## 8. 自定义转换器注册

### 8.1 注册转换器

#### `func RegisterConverter[Src any, Dst any](fn func(Src) (Dst, error)) error`

注册自定义类型转换器，全局生效，优先级高于内置转换器。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fn | `func(Src) (Dst, error)` | 是 | 转换函数，必须返回 error，禁止 panic（内部会 recover） |

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 错误 | `error` | 重复注册同一 (Src,Dst) 返 `ErrDuplicateConverter` |

**约束**：
- 应在 `init()` 或程序启动阶段注册
- 注册后不可注销、不可覆盖
- 并发读安全

**示例**：
```go
type Status int
func (s Status) String() string { ... }

func init() {
    // Status → string
    convertx.RegisterConverter(func(s Status) (string, error) {
        return s.String(), nil
    })
    // string → Status
    convertx.RegisterConverter(func(s string) (Status, error) {
        switch s {
        case "active": return Status(1), nil
        case "inactive": return Status(0), nil
        default: return 0, fmt.Errorf("unknown status: %s", s)
        }
    })
}
```

---

## 9. 选项函数（Option）

所有选项通过 `opts ...Option` 参数传入，可组合使用。

| 选项函数 | 签名 | 说明 |
|---------|------|------|
| `WithDefault` | `func(v any) Option` | 转换失败/零值时返回默认值 |
| `WithStrictMode` | `func() Option` | 结构体/map 转换时未匹配字段返回错误 |
| `WithTagName` | `func(name string) Option` | 指定结构体 tag 名（默认 `"convertx"`） |
| `WithValidation` | `func(rules string) Option` | 转换后自动校验 |
| `WithTimeLayout` | `func(layout string) Option` | 指定时间解析格式（跳过自动识别） |
| `WithTimeUnit` | `func(unit TimeUnit) Option` | 指定时间戳精度（跳过自动推断） |
| `WithDefaultLocation` | `func(loc *time.Location) Option` | 无时区字符串的默认时区（默认 Local） |
| `WithNilPointerAsError` | `func() Option` | nil 指针返回错误而非零值 |

**示例**：
```go
// 组合使用：严格模式 + 自定义tag名 + 校验
err := convertx.Struct(src, dst,
    convertx.WithStrictMode(),
    convertx.WithTagName("json"),
    convertx.WithValidation("required"),
)
```

---

## 10. 错误处理规范

### 10.1 错误提取

所有错误均可通过 `errors.As` 提取为 `*errx.Error`：

```go
_, err := convertx.ToInt("abc")
var errxErr *errx.Error
if errors.As(err, &errxErr) {
    fmt.Println(errxErr.Code())       // "CONVERSION_FAILED"
    fmt.Println(errxErr.Get("source_type"))  // "string"
    fmt.Println(errxErr.Get("target_type"))  // "int"
    fmt.Println(errxErr.Get("field_path"))   // "" (顶层转换)
}
```

### 10.2 错误上下文字段

| 字段 | key | 说明 |
|------|-----|------|
| 源类型 | `source_type` | 如 `string`、`*int`、`map[string]any` |
| 目标类型 | `target_type` | 如 `int`、`time.Time` |
| 字段路径 | `field_path` | 嵌套路径如 `User.Address.City`，顶层为空 |
| 值预览 | `value_preview` | 源值截断至 100 字符 |

### 10.3 错误码与场景对照

| 错误码 | 触发场景 | 可能的 API |
|--------|---------|-----------|
| `ErrConversionFailed` | 类型不兼容、格式非法、nil 指针（默认模式） | 全部转换函数 |
| `ErrValidationFailed` | validx 校验规则不通过 | `ToAndValidate`、`Validate`、`ValidateField` |
| `ErrOverflow` | 数值超出目标类型范围 | `To[int8]` 等数值转换 |
| `ErrUnknownFormat` | 时间字符串不匹配任何格式 | `ToTime` |
| `ErrCircularReference` | 结构体转换检测到循环引用 | `Struct`、`StructToMap` |
| `ErrDuplicateConverter` | 重复注册同一 (源,目标) 转换器 | `RegisterConverter` |
| `ErrNilPointer` | 开启 nil 报错模式时遇到 nil 指针 | 全部转换函数（需选项开启） |
| `ErrUnmatchedField` | 严格模式下有未匹配字段 | `Struct`、`MapToStruct` |

---

## 11. API 速查表

| 类别 | 函数 |
|------|------|
| 基础转换 | `To[T]`、`ToInt`、`ToString`、`ToBool`、`ToFloat64` 等 16 个 |
| 默认值 | `ToOrDefault[T]` |
| 结构体 | `Struct`、`ToStruct[T]`、`StructToMap`、`MapToStruct` |
| 集合 | `SliceToSlice`、`MapToMap`、`SliceToMap`、`MapToSlice` |
| 时间 | `ToTime`、`TimeToString`、`TimestampToTime`、`TimeToTimestamp`、`ToUTC`、`ToLocal`、`ToLocation`、`ToDuration` |
| 校验 | `ToAndValidate[T]`、`Validate`、`ValidateField` |
| 注册 | `RegisterConverter[Src,Dst]` |
| 选项 | `WithDefault`、`WithStrictMode`、`WithTagName`、`WithValidation`、`WithTimeLayout`、`WithTimeUnit`、`WithDefaultLocation`、`WithNilPointerAsError` |
| 常量 | `Version`、8个错误码、`UnitSecond`/`UnitMillisecond`/`UnitMicrosecond`/`UnitNanosecond` |

---

> **文档状态**：API 文档初稿，待人工审核确认后进入阶段8（AI辅助编码开发）。
>
> **审核要点**：函数签名是否合理、参数说明是否完整、错误码是否覆盖所有场景、示例是否正确可运行。
