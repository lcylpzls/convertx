# 贡献指南

感谢参与 convertx 的打磨。请遵循以下约定。

## 环境与语言

- 开发机为 Windows，执行命令一律使用 PowerShell；
- 所有日志、注释、错误信息与文档使用简体中文；
- 目标 Go 版本见 go.mod（当前 1.26.5）。

## 开发流程

1. 从 CHANGELOG 与 docs/ 系列文档中梳理待办；
2. 在分支上实现，代码风格对齐现有文件（薄封装、显式命名、默认安全）；
3. 本地验证：

```powershell
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@2026.1 ./...
go test -race -coverprofile=coverage.out ./...
go tool cover -func coverage.out
```

要求语句覆盖率 100%；新增 API 必须同步更新 README、CHANGELOG 与 API 文档。

## 新增转换能力规范

- 所有导出函数必须返回 `(T, error)` 或 `error`，禁止仅返值变体；
- 错误统一通过 errx 封装（`CONVERTX_` 前缀错误码 + 上下文字段）；
- 禁止 panic：类型断言必须使用安全断言，指针解引用前检查 nil；
- 每个转换函数需覆盖正常 / 边界 / 异常 / nil 四条测试路径；
- 新增选项通过 `Option` 函数注入，不破坏已有 API 签名。

## 提交规范

- 提交信息以版本或主题开头，简述变更与验证结果；
- 涉及行为变更时在 CHANGELOG 记录；
- 版本号基线为 v1.0.0，后续变更在 v1.x.y 内递增，不发布 v2+ 主版本。
