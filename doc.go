// Package convertx 提供 Go 通用类型转换与校验一体化基座：
// 基础类型互转、结构体 / map / 集合 / 时间转换、指针处理、默认值机制、
// 自定义转换器注册、转换后自动校验与独立校验，统一 errx 错误封装，
// 零 panic、零吞错。
// 实现主体位于 internal/core，本包仅暴露稳定公开 API。
package convertx
