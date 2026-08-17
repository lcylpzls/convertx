# 安全说明

convertx 将以下情况视为安全缺陷:

- 任何导出 API 路径 panic(拒绝服务);
- 错误消息泄露敏感字段值(`value_preview` 超出 100 字符截断);
- 自定义转换器 panic 未被捕获或错误链丢失。

## 报告漏洞

- 请勿在公开 issue 中披露漏洞细节;
- 通过邮件 [lcylpzls@qq.com](mailto:lcylpzls@qq.com) 联系维护者,
  并在标题注明 `[Security]`;
- 修复发布前我们会与报告者保持沟通;发布后可应要求公开致谢。

## 安全使用建议

- 转换失败的错误中 `value_preview` 仅保留源值摘要(截断至 100 字符),
  业务侧请勿将完整源值拼入自定义错误消息;
- 自定义转换器由调用方提供,其行为(含 panic)由调用方负责,
  库仅做防御性 recover 并转为错误返回;
- 校验失败路径构造 errx 错误(含调用栈),敏感场景可
  `errx.SetStackCapture(false)` 全局关闭。
