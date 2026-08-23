# QA Record

## Round 1 · 2026-08-23 17:08 GMT+8 · Cost ¥0

执行环境：Docker Compose（backend healthy / frontend healthy）。全程 Mock：`STORAGE_DRIVER=local`，`NOTIFIER_MODE=mock`，SMTP 打到 MailHog。无按量计费调用。

### 结果

| 项 | 结果 | 说明 |
|---|---|---|
| Docker Build | PASS | backend / frontend 多阶段构建成功 |
| Health / Ready | PASS | `/healthz` ok，`/readyz` ready + storage=local |
| API Smoke（容器网络） | PASS | `docker run --network gopuppy_default python:3.12-alpine` 执行 `tests/api_smoke.py` 输出 `SMOKE_OK` |
| 登录 + 年龄到天 | PASS | 奶油 3岁5个月8天 / 共1257天；豆豆 4岁0个月3天 |
| 今日打卡 | PASS | 奶油早/午显示「林爸爸 17:04」 |
| 健康时间轴 | PASS | 绝育 / 狂犬 / 驱虫 / 症状 / 动态重算探针 均可见 |
| 提醒动态重算 | PASS | 补录疫苗后 `next_due_at` 从 2026-06-18 变为 2027-08-23 |
| 账本 Echarts | PASS | 3 个 canvas；本月 ¥233 / 年度 ¥1780 / Top3 非空 |
| 相册空态 | PASS | 「相册还是空白页」+ 上传入口 |
| 跨家庭 404 | PASS | smoke 断言随机 pet UUID 返回 404 |
| 打卡幂等 | PASS | 重复 NOON FEED，活跃记录 ≤ 1 |
| 后端重启恢复 | PASS | restart 后 `/readyz` ready，listen 日志出现第二次 |
| 单测 | PASS | domain / reminder / storage / auth / clock / httputil / notifier / service |
| Playwright 容器 | DEFERRED | `tests/e2e_flow.spec.ts` 与 `qa` profile 已就绪；本轮用浏览器走完同等 5 条路径，避免再拉 1.49.1 镜像 |

### 日志摘要

```
SMOKE_OK
{"status":"ready","storage":"local"}
before 2026-06-18T00:00:00Z
after 2027-08-23T00:00:00Z event 动态重算探针
```

首轮种子失败（checkin SQL 跳号 `$1,$3,$5` 导致 42P18）已在本轮修复，未更换方案。
