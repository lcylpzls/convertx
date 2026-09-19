# 更新日志

本项目遵循[语义化版本](https://semver.org/lang/zh-CN/)。

## [v1.0.1] - 2026-09-19

### 修复

- 修复模块无法被拉取：`docs/` 下中文文件名导致 Go 模块 zip 打包失败
  （`malformed file path ... invalid char '（'`），v1.0.0 在 proxy.golang.org 与
  sum.golang.org 上始终 404，`go get github.com/lcylpzls/convertx` 不可用；
- 新增 `.gitattributes` 将 `docs/` 排除出模块归档，模块恢复可拉取。

## [v1.0.0] - 2026-08-17

### 新增

- 基础类型互转：string / 数值 / bool / byte / rune 全量互转，泛型入口 `To[T]` 与类型特定函数；
- 结构体转换：struct ↔ struct、struct ↔ map，支持 tag 映射、匿名字段展开、嵌套递归、循环引用检测与严格模式；
- 集合类型转换：slice ↔ slice、map ↔ map、slice ↔ map、map ↔ slice，单元素失败整体报错；
- 时间类型转换：string ↔ time.Time（多格式自动识别）、时间戳精度自动推断、时区转换、Duration 互转；
- 指针处理：任意层级指针解引用与重建，nil 安全；
- 默认值机制：`ToOrDefault` 与 `WithDefault` 选项；
- 自定义转换器注册：`RegisterConverter[Src, Dst]`，优先级高于内置；
- 校验集成：`ToAndValidate`、`Validate`、`ValidateField`，错误统一 errx 封装；
- 统一错误封装：8 个 `CONVERTX_` 前缀错误码，携带 source_type / target_type / field_path / value_preview 上下文。

### 修复

- 根包补齐 `With*` 系列选项构造函数，与文档公开 API 对齐；
- 接口目标转换增加实现检查，避免嵌套字段赋值 panic，并保留指针方法集；
- 修复 `SliceToSlice` / `MapToMap` 使用接口泛型参数时的 nil 反射类型 panic；
- 修复 nil 指针嵌入字段在 `Struct` / `MapToStruct` / `StructToMap` 中的 panic；
- 修复 `StructToMap` 遇到 `time.Time` 字段时转换失败的问题；
- `WithDefault` 支持零值触发、默认值类型转换，并补齐 `Struct` / `MapToStruct` 的默认值与校验选项语义。
