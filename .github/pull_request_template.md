## 变更说明

（简述本次变更解决了什么问题、如何实现）

## 关联项

- Closes #
- 对应文档:`docs/07-api-reference.md` 中的 API 或 docs/architecture.md 中的设计

## 验证

- [ ] `go vet ./...` 通过
- [ ] `staticcheck ./...` 通过
- [ ] `go test -race -coverprofile=coverage.out ./...` 通过,覆盖率 100%
- [ ] 新增转换函数时正常 / 边界 / 异常 / nil 四路径测试齐全
- [ ] 新增错误场景时错误码(`CONVERTX_` 前缀)与上下文字段断言齐全
- [ ] 涉及 API 变更时同步更新 README、CHANGELOG 与 API 文档

## 兼容性

（是否破坏现有 API;V1 大版本内只增不改;涉及公共行为变更需在 PR 说明并在 CHANGELOG 记录）
