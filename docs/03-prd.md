# 项目 PRD — convertx 通用类型转换与校验一体化库

| 项目 | convertx |
|------|----------|
| 文档版本 | v1.0 |
| 编制日期 | 2026-08-17 |
| 前置文档 | 01-project-charter.md、02-requirements-checklist.md |

---

## 1. 文档概述

### 1.1 目的

本文档为 convertx 库的正式产品需求基线，是开发、测试、文档编写的唯一依据。所有功能行为、异常处理、性能指标均以本文档为准。

### 1.2 范围

覆盖 V1 版本全部核心能力（C01-C11）及兼容需求（K01-K03）。V2 高性能能力（S01-S02）仅作预留说明，不纳入 V1 验收。

### 1.3 术语

| 术语 | 定义 |
|------|------|
| 源类型（Source Type） | 转换输入值的 Go 类型 |
| 目标类型（Target Type） | 转换期望输出的 Go 类型 |
| 字段路径（Field Path） | 结构体嵌套字段的点分路径，如 `User.Address.City` |
| 转换器（Converter） | 将特定源类型转为目标类型的函数，签名为 `func(Src) (Dst, error)` |
| 校验规则（Validation Rule） | validx 规则字符串，如 `"min=0,max=100"` |

---

## 2. 用户角色

| 角色 | 描述 | 核心诉求 |
|------|------|---------|
| **应用开发者** | 在业务项目中引入 convertx 的 Go 开发者 | API 简洁易用、错误信息清晰、转换行为可预期 |
| **架构师 / 基础库维护者** | 负责家族库选型和规范制定 | 依赖可控、零 panic、统一错误封装、与 errx/validx 无缝集成 |
| **开源贡献者** | 向 convertx 提交 PR 的社区开发者 | 代码结构清晰、测试规范、贡献指南明确 |

---

## 3. 整体业务说明

### 3.1 库定位

convertx 是 Go 语言的**通用类型转换与校验一体化基础库**，解决以下核心问题：

1. **碎片化**：项目中同时引入 cast、mapstructure、copier、validator 多库，API 和错误风格不统一
2. **吞错风险**：cast 的 `ToXxx()` 变体静默返零值，开发者易误用
3. **转换校验割裂**：转换后需手动调用校验库，样板代码重复
4. **指针/嵌套繁琐**：现有库对多级指针、嵌套结构体指针处理不统一

### 3.2 典型使用场景

| 场景 | 描述 | 涉及能力 |
|------|------|---------|
| 配置归一化 | YAML/JSON 解析为 `map[string]any` 后转强类型 struct | C03、C02、C09 |
| API 入参处理 | HTTP query/header 中的 string 转 int/time 并校验范围 | C01、C05、C09 |
| 模型转换 | DO ↔ DTO ↔ VO 结构体字段复制与类型转换 | C02、C06 |
| 数据库可空字段 | `sql.NullString` 等与业务模型互转 | C01、C06 |
| 批量数据转换 | `[]string` 转 `[]int`、map 值类型批量转换 | C04 |
| 时间处理 | 字符串时间解析、时间戳转换、时区转换 | C05 |

---

## 4. 核心能力详述

### 4.1 基础类型互转（C01）

#### 4.1.1 能力说明

提供 Go 所有基础数值类型、string、bool、byte、rune 之间的安全互转。转换失败返回 `errx.Error`，不 panic、不返零值吞错。

#### 4.1.2 API 形态

- 泛型入口：`convertx.To[Dst](src any) (Dst, error)`
- 类型特定函数：`convertx.ToInt(src any) (int, error)`、`convertx.ToString(src any) (string, error)` 等
- 带默认值：`convertx.ToOrDefault[Dst](src any, def Dst) Dst`

#### 4.1.3 转换规则

| 源 → 目标 | 规则 |
|-----------|------|
| 数值 → 数值 | 直接类型转换，溢出时返回 `ErrOverflow` 错误 |
| string → 数值 | 使用 `strconv` 解析，失败返回 `ErrConversionFailed` |
| 数值 → string | `fmt.Sprintf("%v")` |
| string → bool | 识别 `"true"/"false"/"1"/"0"/"t"/"f"`（大小写不敏感），其他返回错误 |
| bool → string | `"true"/"false"` |
| bool → 数值 | true→1, false→0 |
| 数值 → bool | 非零→true, 零→false |
| byte/rune | 按 uint8/int32 处理 |

#### 4.1.4 约束

- 所有函数返回 `(T, error)`，不提供仅返 T 的变体
- 溢出检测：int8 最大值 127，源值 128 转 int8 返回错误
- string 前后空白自动 trim 后再解析

---

### 4.2 结构体转换（C02）

#### 4.2.1 能力说明

在不同 struct 类型之间按字段名或 tag 进行值复制，字段类型不同时自动调用基础类型转换器。支持嵌套结构体递归转换。

#### 4.2.2 API 形态

- `convertx.Struct(src any, dst any, opts ...Option) error`
- 泛型版本：`convertx.ToStruct[Dst](src any, opts ...Option) (Dst, error)`

#### 4.2.3 字段映射规则

1. 默认按字段名匹配（大小写不敏感？需确认——本 PRD 默认**大小写敏感**，与 Go 导出字段惯例一致）
2. 源字段 tag `convertx:"target_name"` 指定目标字段名
3. `convertx:"-"` 忽略该字段
4. 匿名字段（embedded struct）自动展开，字段提升
5. 未匹配字段：默认忽略，`WithStrictMode()` 选项开启后返回错误

#### 4.2.4 约束

- 目标必须为非 nil 指针
- 源为 nil 指针时，目标各字段设为零值
- 循环引用检测：转换过程中维护已访问指针集合，检测到循环返回 `ErrCircularReference`
- 字段类型不同时调用 C01 基础转换；基础转换失败则整体失败并携带 field_path

---

### 4.3 结构体与 map 互转（C03）

#### 4.3.1 能力说明

struct → `map[string]any`：将结构体字段导出为 map；`map[string]any` → struct：将 map 键值对填充到结构体字段。

#### 4.3.2 API 形态

- `convertx.StructToMap(src any, opts ...Option) (map[string]any, error)`
- `convertx.MapToStruct(src map[string]any, dst any, opts ...Option) error`
- 支持强类型 map：`map[string]string` → struct 等

#### 4.3.3 规则

- map key 默认使用字段名，可通过 tag `convertx:"json_name"` 自定义
- 嵌套结构体转为嵌套 map
- map → struct 时多余 key 默认忽略，严格模式报错
- 值类型不匹配时调用基础转换器

---

### 4.4 集合类型转换（C04）

#### 4.4.1 能力说明

slice 和 map 在元素/键值类型不同时的批量转换，以及 slice ↔ map 互转。

#### 4.4.2 API 形态

- `convertx.SliceToSlice[Src, Dst](src []Src) ([]Dst, error)`
- `convertx.MapToMap[SK, SV, DK, DV](src map[SK]SV) (map[DK]DV, error)`
- `convertx.SliceToMap[T, K](src []T, keyFn func(T) K) (map[K]T, error)`
- `convertx.MapToSlice[K, V](src map[K]V) ([]V, error)`（顺序不保证）

#### 4.4.3 约束

- 任一元素转换失败则整体返回错误，**不返回部分结果**
- nil slice 输入返回 nil slice（非空）
- nil map 输入返回 nil map
- slice 转换保留原顺序
- map→slice 的元素顺序不保证（Go map 迭代无序）

---

### 4.5 时间类型转换（C05）

#### 4.5.1 能力说明

string、int64（时间戳）、`time.Time`、`time.Duration` 之间的互转，支持多格式自动识别和时区转换。

#### 4.5.2 API 形态

- `convertx.ToTime(src any, opts ...TimeOption) (time.Time, error)`
- `convertx.TimeToString(t time.Time, layout string) string`
- `convertx.TimestampToTime(ts int64, unit TimeUnit) time.Time`
- `convertx.TimeToTimestamp(t time.Time, unit TimeUnit) int64`
- `convertx.ToDuration(src any) (time.Duration, error)`

#### 4.5.3 格式识别（string → time）

按优先级尝试以下格式，首个成功即返回：

1. RFC3339 / RFC3339Nano（带时区偏移，如 `2026-08-17T10:00:00+08:00`）
2. `2006-01-02 15:04:05`（无时区，默认 Local）
3. `2006-01-02`（日期，零点）
4. RFC822 / RFC822Z
5. 可通过 `WithTimeLayout(layout)` 指定格式，指定后仅尝试该格式

#### 4.5.4 时间戳精度识别（int64 → time）

按数值范围自动推断：

| 范围 | 精度 |
|------|------|
| < 1e11 | 秒（s） |
| 1e11 ~ 1e14 | 毫秒（ms） |
| 1e14 ~ 1e17 | 微秒（μs） |
| ≥ 1e17 | 纳秒（ns） |

可通过 `WithTimeUnit(unit)` 强制指定精度，跳过自动推断。

#### 4.5.5 时区处理

- string 带时区偏移：解析后保留该时区
- string 不带时区：默认 Local，可通过 `WithDefaultLocation(time.UTC)` 改为 UTC
- `time.Time` 转 string：默认使用值自身时区
- 提供 `convertx.ToUTC(t time.Time) time.Time`、`convertx.ToLocal(t time.Time) time.Time` 快捷转换

---

### 4.6 指针处理（C06）

#### 4.6.1 能力说明

任意层级指针的解引用与重建，nil 安全。

#### 4.6.2 规则

| 操作 | 规则 |
|------|------|
| `*T` → `T` | 解引用；nil 指针返零值（可配置 `WithNilPointerAsError()` 改为返错误） |
| `T` → `*T` | 取地址返回新指针 |
| `**T` → `T` | 递归解引用至非指针 |
| `T` → `**T` | 递归取地址 |
| `*T` → `*U` | 解引用 → 基础转换 → 取地址 |

#### 4.6.3 约束

- 任何层级 nil 指针不 panic
- 结构体转换中指针字段按上述规则自动处理
- 取地址操作返回新分配的指针，不共享源值地址

---

### 4.7 默认值机制（C07）

#### 4.7.1 能力说明

转换失败或源为零值/nil 时返回调用方指定的默认值。

#### 4.7.2 API 形态

- `convertx.ToOrDefault[Dst](src any, def Dst) Dst`
- 选项模式：`convertx.To[Dst](src, convertx.WithDefault(def)) (Dst, error)` — 此模式下永远不返 error

#### 4.7.3 触发条件

默认值在以下情况生效：
1. 源值为 nil（指针/接口/map/slice/function）
2. 源值为类型零值（`""`、`0`、`false`）
3. 转换过程发生错误（溢出、格式非法等）

> **注意**：源为零值时触发默认值意味着 `convertx.ToIntOrDefault("0", -1)` 返回 `-1` 而非 `0`。这是设计决策，如需区分"零值"和"缺失"，应使用指针类型或 C01 标准 API 自行判断。

---

### 4.8 自定义转换函数注册（C08）

#### 4.8.1 能力说明

允许注册业务自定义类型的转换器，覆盖或补充内置转换逻辑。

#### 4.8.2 API 形态

- `convertx.RegisterConverter[Src, Dst](fn func(Src) (Dst, error)) error`
- 内部以 `(reflect.TypeOf(Src), reflect.TypeOf(Dst))` 为唯一键

#### 4.8.3 规则

1. 自定义转换器优先级**高于**内置转换器
2. 重复注册同一 (源类型, 目标类型) 返回 `ErrDuplicateConverter`
3. 注册应在 `init()` 或程序启动阶段完成，运行时只读
4. 转换器内部必须返回 error，禁止 panic；panic 会被 recover 并转为 `ErrConversionFailed`（防御性）
5. 转换器返回的 error 会被 `errx.Wrap` 封装，保留原始错误链

#### 4.8.4 约束

- 全局单例注册表，并发读安全
- 不支持注销（注册后全局生效）
- 泛型注册函数在编译期确定 Src/Dst 类型

---

### 4.9 转换后自动校验（C09）

#### 4.9.1 能力说明

转换成功后自动对目标值执行 validx 校验规则，转换+校验一体化。

#### 4.9.2 API 形态

- `convertx.ToAndValidate[Dst](src any, rules string) (Dst, error)`
- 选项模式：`convertx.To[Dst](src, convertx.WithValidation(rules)) (Dst, error)`

#### 4.9.3 执行流程

```
源值 → 类型转换 → 转换失败？→ 是：返回转换错误（不执行校验）
                  → 否：执行 validx 校验 → 校验失败？→ 是：返回校验错误
                                                  → 否：返回转换值
```

#### 4.9.4 约束

- 校验规则语法与 validx 完全一致（如 `"min=0,max=100"`、`"required,email"`）
- 校验失败返回的错误通过 errx 封装，Kind 为 `ValidationFailed`
- 结构体转换时可对整个结构体或单个字段指定校验规则

---

### 4.10 独立校验快捷 API（C10）

#### 4.10.1 能力说明

对已有强类型值直接执行校验，无需转换。

#### 4.10.2 API 形态

- `convertx.Validate(value any, rules string) error`
- `convertx.ValidateField(value any, rules string) error`

#### 4.10.3 约束

- 内部委托 `validx.Validate` / `validx.ValidateField`
- 错误通过 errx 封装（validx 原生错误转为 errx 格式）
- 规则语法、行为与 validx 完全一致，本库不做语义修改

---

### 4.11 统一错误封装（C11）

#### 4.11.1 错误码定义

| 错误码常量 | 含义 | 触发场景 |
|-----------|------|---------|
| `ErrConversionFailed` | 通用转换失败 | 类型不兼容、格式非法等 |
| `ErrValidationFailed` | 校验失败 | validx 校验规则不通过 |
| `ErrOverflow` | 数值溢出 | 源值超出目标类型范围 |
| `ErrUnknownFormat` | 时间格式无法识别 | string 不匹配任何已知时间格式 |
| `ErrCircularReference` | 循环引用 | 结构体转换中检测到循环引用 |
| `ErrDuplicateConverter` | 转换器重复注册 | 同一 (源,目标) 类型已注册 |
| `ErrNilPointer` | 空指针 | 开启 nil 指针报错模式时遇到 nil |
| `ErrUnmatchedField` | 未匹配字段 | 严格模式下结构体/map 有未匹配字段 |

#### 4.11.2 错误上下文字段

所有错误必须携带以下字段（通过 `errx.WithKV`）：

| 字段 | 说明 |
|------|------|
| `source_type` | 源值的 Go 类型字符串 |
| `target_type` | 目标类型字符串 |
| `field_path` | 结构体/map 中的字段路径（顶层转换时为空） |
| `value_preview` | 源值的摘要（截断至 100 字符，避免敏感数据泄露） |

#### 4.11.3 约束

- 所有导出函数返回的 error 均可通过 `errors.As(err, &errxErr)` 提取为 `*errx.Error`
- 禁止 panic：内部任何可能 panic 的路径（类型断言、nil 解引用、数组越界）必须 recover 并转为错误
- 禁止吞 error：不提供仅返值不返 error 的 API

---

## 5. 业务流程

### 5.1 标准转换流程

```
调用方传入 (src, 目标类型, 选项)
    │
    ▼
参数校验（目标类型有效性、非 nil 指针等）
    │
    ▼
查找转换器：自定义注册表 → 内置转换器
    │
    ▼
执行转换
    │
    ├─ 失败 → 封装 errx.Error（含 source_type/target_type/field_path）→ 返回
    │
    ▼
转换成功
    │
    ├─ 有校验选项？→ 执行 validx 校验 → 失败返回 ErrValidationFailed
    │                              → 成功继续
    │
    ▼
返回 (目标值, nil)
```

### 5.2 结构体转换流程

```
传入 (src struct, dst *struct, opts)
    │
    ▼
反射解析源结构体字段元信息（V2 缓存命中则跳过）
    │
    ▼
遍历源字段：
    ├─ 字段 tag = "-" → 跳过
    ├─ 按 tag/字段名匹配目标字段 → 未匹配？
    │   ├─ 严格模式 → 记录 ErrUnmatchedField
    │   └─ 默认模式 → 跳过
    ├─ 字段类型相同 → 直接赋值
    ├─ 字段类型不同 → 递归调用标准转换流程
    └─ 嵌套结构体 → 递归执行本流程
    │
    ▼
循环引用检测（已访问指针集合）
    │
    ▼
返回 error 或 nil
```

### 5.3 自定义转换器注册流程

```
init() 阶段调用 RegisterConverter[Src, Dst](fn)
    │
    ▼
检查 (Src类型, Dst类型) 是否已注册
    ├─ 已注册 → 返回 ErrDuplicateConverter
    └─ 未注册 → 存入全局注册表 → 返回 nil
    │
    ▼
运行时转换时优先查找自定义注册表
```

---

## 6. 异常场景

| 场景 | 处理方式 | 错误码 |
|------|---------|--------|
| 源值为 nil 且目标为非指针 | 返回零值 + error（或默认值，取决于选项） | `ErrConversionFailed` |
| 数值转换溢出 | 返回零值 + error | `ErrOverflow` |
| string 转数值格式非法 | 返回零值 + error | `ErrConversionFailed` |
| string 转时间无匹配格式 | 返回零值 + error | `ErrUnknownFormat` |
| 结构体目标为 nil 指针 | 返回 error | `ErrConversionFailed` |
| 结构体循环引用 | 返回 error，不栈溢出 | `ErrCircularReference` |
| map→struct 严格模式下多余 key | 返回 error，列出未匹配 key | `ErrUnmatchedField` |
| 集合转换单个元素失败 | 整体返回 error，无部分结果 | 元素对应的错误码 |
| 自定义转换器 panic | recover 后转为错误 | `ErrConversionFailed` |
| 校验规则语法错误 | 返回 error | `ErrValidationFailed` |
| 重复注册转换器 | 返回 error | `ErrDuplicateConverter` |

---

## 7. 状态流转

### 7.1 转换器注册表状态

```
未初始化 ──init()注册──→ 已注册（运行时只读）
     │                      │
     └──运行时注册──→ 不推荐但允许（并发安全）
```

- 注册表一旦写入，运行期不可修改、不可注销
- 并发读安全（`sync.RWMutex` 或 `sync.Map`）

### 7.2 转换结果状态

```
待转换 ──转换成功──→ 已转换 ──校验通过──→ 完成
                  │              │
                  │              └─校验失败──→ 校验错误
                  └─转换失败──→ 转换错误
```

---

## 8. 限制约束

### 8.1 功能限制

| 限制项 | 说明 |
|--------|------|
| 不做序列化 | 不提供 JSON/XML/YAML 编解码，仅处理已解析为 Go 值的转换 |
| 不做 ORM | 不涉及数据库连接、SQL、表映射 |
| 不做深拷贝 | 结构体转换为字段值复制，引用类型字段（slice/map/pointer）共享底层数据 |
| 不做表达式查询 | 不支持 JMESPath/jq 等路径表达式 |
| 无 panic 变体 | 不提供 `MustXxx()` 系列 |
| 无吞错变体 | 不提供 `ToXxx()` 仅返值系列 |

### 8.2 性能限制（V1）

- V1 基于反射实现，性能与 `mapstructure`、`cast` 同量级（±15%）
- 结构体转换首次调用需反射解析，有一次性开销
- 高频场景（>10万次/秒）建议等待 V2 零反射路径

### 8.3 安全限制

- 错误信息中 `value_preview` 截断至 100 字符，避免大值或敏感数据完整泄露
- 不执行源值中的任何代码或表达式
- 自定义转换器由调用方提供，其安全性由调用方负责

---

## 9. 性能要求

| 指标 | V1 目标 | V2 目标 |
|------|---------|---------|
| 基础类型转换（string→int） | 与 `spf13/cast` 持平（±10%） | 优于 cast 20%+ |
| 结构体转换（10字段平结构体） | 与 `mapstructure` 持平（±15%） | 优于 mapstructure 30%+ |
| 集合转换（1000元素 slice） | 线性时间，无意外开销 | 同左 + 快速路径 |
| 内存分配 | 每次转换 ≤ 目标值大小 + 固定开销 | 减少反射相关分配 |
| 并发安全 | 导出 API 无共享可变状态，并发安全 | 同左 + 缓存并发安全 |

> 性能指标通过 `go test -bench` 基准测试验证，测试用例与竞品对齐。

---

## 10. 兼容性要求

### 10.1 Go 版本

- 最低版本：Go 1.26.5
- 使用特性：泛型、`any`、`errors.Is/As`、`reflect` 标准 API
- CI 覆盖：1.26.5 及最新稳定版

### 10.2 平台与架构

- 操作系统：Linux、macOS、Windows
- 架构：amd64、arm64
- 不使用平台特定 API（`syscall`、`unsafe` 除外的标准库）

### 10.3 依赖兼容

| 依赖 | 最低版本 | 角色 |
|------|---------|------|
| `github.com/lcylpzls/errx` | v1.5.7 | 运行时 |
| `github.com/lcylpzls/validx` | v1.2.5 | 运行时 |
| `github.com/lcylpzls/testx` | v1.4.5 | 测试 |

- 不引入任何非家族第三方依赖
- `go.mod` 中锁定具体版本，不使用 `latest`

### 10.4 API 兼容承诺

- V1 大版本内（v1.x.x）：只增不改，已有函数签名和行为保持向后兼容
- V2 大版本：新增高性能 API，V1 API 完全保留且行为不变
- 废弃 API 提前两个小版本标注 `// Deprecated:`

---

## 11. 验收标准总览

| 维度 | 标准 |
|------|------|
| 功能完整 | C01-C11 全部实现，每条需求的验收标准通过 |
| 测试覆盖 | 单元测试覆盖率 ≥ 90%，每个函数含正常/边界/异常/nil 用例 |
| 错误规范 | 所有错误为 errx 类型，携带规定上下文字段，无 panic、无吞错 |
| 依赖纯净 | `go mod graph` 仅含家族库和标准库 |
| 跨平台 | CI 三平台两架构全部通过 |
| 文档 | README、GoDoc、示例代码完整 |
| 性能 | Benchmark 达标且与竞品对比数据归档 |

---

> **文档状态**：PRD 初稿，待人工审核确认后进入阶段4（开发计划与任务拆解）。
>
> **审核要点**：API 形态是否合理、异常场景是否遗漏、性能指标是否可接受、兼容性约束是否可执行。
