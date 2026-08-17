# convertx 架构设计

> 本文档描述 convertx 的实现架构，是 docs/05-系统架构设计文档.md 的落地说明。

## 1. 整体架构

convertx 为**无状态基础库**：全部 API 为纯函数调用，无后台 goroutine、无外部服务依赖。
唯一全局状态是自定义转换器注册表（init 阶段写入后运行时只读，`sync.RWMutex` 保护）。

```
┌────────────────────────────────────────────┐
│  convertx 根包（公开 API 层）                │
│  api.go — 类型别名 + 函数转发，无业务逻辑      │
├────────────────────────────────────────────┤
│  internal/core（实现层）                    │
│  ┌─────────── 调度层 ───────────┐           │
│  │ convert.go  核心调度器         │           │
│  │ options.go  选项解析          │           │
│  │ registry.go 转换器注册表      │           │
│  ├─────────── 实现层 ───────────┤           │
│  │ scalar.go    基础类型转换     │           │
│  │ struct.go    结构体转换       │           │
│  │ map.go       struct↔map      │           │
│  │ collection.go 集合转换        │           │
│  │ time.go      时间转换         │           │
│  │ pointer.go   指针处理         │           │
│  │ validate.go  校验集成         │           │
│  ├─────────── 基础设施层 ────────┤           │
│  │ errors.go    错误码定义       │           │
│  │ error_util.go errx 封装       │           │
│  │ default.go   默认值机制       │           │
│  │ struct_meta.go 字段元信息     │           │
│  └──────────────────────────────┘           │
└────────────┬────────────────────────────────┘
             ▼
     errx / validx / 标准库
```

## 2. 转换流程

```
Convert(src, dstType, opts...)
  │
  ├─ dstType == nil（目标为 any）→ 直接返回 src
  ├─ src == nil → 默认值 / nil 指针 / 零值+错误
  ├─ 源链含 nil 指针/接口 → 默认值 / ErrNilPointer / 零值
  │
  ▼
convert(srcVal, dstType, cfg, fieldPath)
  ├─ 解引用源指针与接口（递归，记录指针地址做循环引用检测）
  ├─ 计算目标指针层数（解出元素类型）
  ├─ 元素级类型匹配 → 直接返回（值复制）
  ├─ 自定义转换器查找（(srcType, dstType) 精确匹配）
  ├─ 内置转换分发（按 dstType.Kind）
  └─ 重建目标指针层（取地址）
  │
  ▼
转换后校验（WithValidation）→ 失败返回 ErrValidationFailed
```

### 2.1 循环引用检测

结构体转换的指针字段递归时，在 `cfg.path`（`map[uintptr]bool`）中记录
当前递归路径上已解引用的指针地址；重复出现即检测到环，返回
`ErrCircularReference`。递归返回时移除记录，因此**兄弟字段共享同一指针不会误报**。

### 2.2 指针语义

- `*T → T`：解引用；nil 指针默认返回目标零值，`WithNilPointerAsError` 时返回错误。
- `T → *T`：取地址（新分配，不共享源地址）。
- 多级指针：递归解引用/取地址。
- 结构体字段为指针时按上述规则自动处理。

## 3. 转换器注册表

```
RegisterConverter[Src, Dst](fn)
  → key = (reflect.TypeOf(Src), reflect.TypeOf(Dst))
  → 已存在？ErrDuplicateConverter : 存入全局表

运行时 Lookup(srcType, dstType)：读锁查找，自定义优先级高于内置。
```

- 注册表全局生效，注册后不可注销、不可覆盖。
- 执行转换器时防御 panic（recover 转为 `ErrConversionFailed`）。
- 转换器返回的错误经 `errx.WrapCode` 封装，保留原始错误链。

## 4. 错误模型

所有错误均为 `*errx.Error`（可通过 `errors.As` 提取），携带统一上下文：

| 字段 | 说明 |
|------|------|
| `source_type` | 源类型字符串 |
| `target_type` | 目标类型字符串 |
| `field_path` | 嵌套字段路径（如 `User.Address.City`，顶层为空字符串） |
| `value_preview` | 源值摘要（截断至 100 字符） |

8 个错误码统一 `CONVERTX_` 前缀，init 阶段注册（`errx.RegisterCode` +
`RegisterCodeKind`）：

- `CONVERTX_CONVERSION_FAILED`（KindInvalid）
- `CONVERTX_VALIDATION_FAILED`（KindInvalid）
- `CONVERTX_OVERFLOW`（KindInvalid）
- `CONVERTX_UNKNOWN_FORMAT`（KindInvalid）
- `CONVERTX_CIRCULAR_REFERENCE`（KindInvalid）
- `CONVERTX_DUPLICATE_CONVERTER`（KindAlreadyExists）
- `CONVERTX_NIL_POINTER`（KindInvalid）
- `CONVERTX_UNMATCHED_FIELD`（KindInvalid）

校验失败错误：委托 `validx.ValidateField` 后重新封装为
`CONVERTX_VALIDATION_FAILED`，`Unwrap` 保留 validx 原始错误链。

## 5. 时间转换

- string → time.Time：按优先级自动识别 RFC3339(Nano) / `2006-01-02 15:04:05` /
  `2006-01-02` / RFC822(Z)；`WithTimeLayout` 指定后仅尝试该格式。
- 无时区字符串默认 `time.Local`，`WithDefaultLocation` 可改为 UTC 等。
- 整数时间戳：按数值范围自动推断精度（<1e11 秒 / <1e14 毫秒 / <1e17 微秒 / 其余纳秒），
  `WithTimeUnit` 强制指定。

## 6. 并发安全

- 导出 API 无共享可变状态（每次调用新建 `convertConfig`）。
- 注册表 `sync.RWMutex` 保护，运行时只读、并发读安全。
- `-race` 测试通过。
