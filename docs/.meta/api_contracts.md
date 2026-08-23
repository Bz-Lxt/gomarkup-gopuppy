# API Contract Gate · 2026-08-23 16:20 GMT+8

| Provider | 验证方式 | 状态 | 备注 |
|---|---|---|---|
| SMTP | 真实 SMTP 拨号至 MailHog (`mailhog:1025`) | **verified (local)** | 走标准 `net/smtp`，非伪造发送函数 |
| 企业微信群机器人 | 无真实 webhook URL | **UNVERIFIED** | `HTTPSender` 真实 HTTP POST 路径已接线；`NOTIFIER_MODE=mock` 或 URL 为空时落入 `mock_deliveries` |
| 通用 Webhook | 同上 | **UNVERIFIED** | 同上 |
| 阿里云 OSS | 无密钥 | **UNVERIFIED** | `storage.Remote` PUT/GET/DELETE 已实现；缺凭证返回明确配置错误 |
| 腾讯云 COS | 无密钥 | **UNVERIFIED** | 同上 |

切换开关见最终 `README.md` §7。默认 `STORAGE_DRIVER=local`、`NOTIFIER_MODE=mock`。
