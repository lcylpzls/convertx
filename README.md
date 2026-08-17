# convertx

Go 通用类型转换与校验一体化基座：基础类型、结构体、集合、时间的类型转换，
转换后自动校验与独立校验，统一 errx 错误封装，零 panic、零吞错。

> 当前状态：**v1.0.0**（已发布）。

## 定位

convertx 解决类型转换场景中多库拼装、错误处理不规范、转换与校验割裂的问题：

- **基础类型互转**：string / 数值 / bool / byte / rune 全量互转，溢出检测、错误明确；
- **结构体转换**：struct ↔ struct、struct ↔ map，tag 映射、匿名字段展开、嵌套递归、循环引用检测；
- **集合转换**：slice / map 批量转换，单元素失败整体报错，nil 安全；
- **时间转换**：多格式自动识别、时间戳精度自动推断、时区转换、Duration 互转；
- **转换 + 校验一体化**：转换成功后自动执行 validx 规则；
- **统一错误**：所有错误为 `*errx.Error`，携带 source_type / target_type / field_path 上下文。

## 快速上手

```go
import "github.com/lcylpzls/convertx"

// 基础类型转换
v, err := convertx.ToInt("42") // v=42, err=nil

// 带默认值（失败/零值/nil 时返回默认值）
v := convertx.ToOrDefault[int]("abc", -1) // v=-1

// 转换后自动校验
v, err := convertx.ToAndValidate[int]("150", "min=0,max=100") // err=ErrValidationFailed

// 结构体转换（字段类型不同自动转换）
type DO struct{ Name string; Age int }
type VO struct{ Name string; Age string }
var vo VO
err := convertx.Struct(&do, &vo)

// 自定义转换器（优先级高于内置）
convertx.RegisterConverter(func(s Status) (string, error) { ... })

// 时间解析（自动识别格式）
t, err := convertx.ToTime("2026-08-17T10:00:00+08:00")
```

## API 索引

| 类别 | API |
|------|-----|
| 基础转换 | `To[T]`、`ToInt`、`ToInt8`、`ToInt16`、`ToInt32`、`ToInt64`、`ToUint`、`ToUint8`、`ToUint16`、`ToUint32`、`ToUint64`、`ToFloat32`、`ToFloat64`、`ToBool`、`ToString`、`ToByte`、`ToRune` |
| 默认值 | `ToOrDefault[T]` |
| 结构体 | `Struct`、`ToStruct[T]`、`StructToMap`、`MapToStruct` |
| 集合 | `SliceToSlice`、`MapToMap`、`SliceToMap`、`MapToSlice` |
| 时间 | `ToTime`、`TimeToString`、`TimestampToTime`、`TimeToTimestamp`、`ToUTC`、`ToLocal`、`ToLocation`、`ToDuration` |
| 校验 | `ToAndValidate[T]`、`Validate`、`ValidateField` |
| 注册 | `RegisterConverter[Src, Dst]` |
| 选项 | `WithDefault`、`WithStrictMode`、`WithTagName`、`WithValidation`、`WithTimeLayout`、`WithTimeUnit`、`WithDefaultLocation`、`WithNilPointerAsError` |

## 错误处理

所有错误均为 `*errx.Error`，可通过 `errors.As` 提取：

```go
_, err := convertx.ToInt("abc")
var e *errx.Error
if errors.As(err, &e) {
    fmt.Println(e.Code()) // "CONVERTX_CONVERSION_FAILED"
    for _, kv := range e.Fields() {
        fmt.Printf("%s=%v\n", kv.Key, kv.Value) // source_type / target_type / field_path / value_preview
    }
}
```

错误码：`ErrConversionFailed`、`ErrValidationFailed`、`ErrOverflow`、
`ErrUnknownFormat`、`ErrCircularReference`、`ErrDuplicateConverter`、
`ErrNilPointer`、`ErrUnmatchedField`。

## 安装

```powershell
go get github.com/lcylpzls/convertx
```

Go 版本要求：1.26.5+。

## 质量

- 单元测试语句覆盖率 100%（convertx 与 internal/core 双包）；
- race / vet / staticcheck / fuzz 全绿；
- 依赖仅限家族库（errx / validx / testx）与标准库。

## 文档

- [架构设计](docs/architecture.md)
- [API 接口文档](docs/07-API接口文档.md)
- [示例](examples/basic/)

## 贡献

见 [CONTRIBUTING.md](CONTRIBUTING.md)。
