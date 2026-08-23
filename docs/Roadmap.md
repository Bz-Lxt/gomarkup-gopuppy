# GoPuppy · 实施路线图（WHEN）

> **权威声明**：本文件定义 **WHEN**。`docs/Requirements.md` 定义 **WHAT**。
> 规模：Tier 2（10k–15k LoC）→ MVP / V1 / V2 边界强制冻结。

| 项 | 值 |
|---|---|
| 冻结时间 | 2026-08-23 16:07 (GMT+8) |
| Phase Order | **Logic-First**（交换 Phase 2 / Phase 3） |
| 开发端口 | 27151–27156（15231 段被同机项目占用后改派） |

---

## Phase Order Decision

**裁定：Logic-First。**

理由：健康时间轴与 Echarts 体重/开销图的组件结构直接派生于冻结枚举与按月聚合契约；先建 schema + API，UI 才能对着真实形状而不是臆造字段渲染。

执行顺序：Phase 1 Architect → Phase 3 Logic → Phase 2 UI → Phase 4 QA → Phase 5 Auditor。

---

## 分期边界

### MVP（本轮 `/auto` 必须交付）

账户家庭 RBAC、宠物档案（年龄精确到天）、今日打卡 + WebSocket 同家庭同步、健康时间轴（含症状）、提醒引擎（周期计算 / 动态重算 / 定时扫描 / 三通道 / 幂等 / 窄重试 / 提前提醒 / 通知中心）、记账与体重曲线（后端聚合 + Echarts）、媒体上传（local 默认 + OSS/COS 驱动落地）、种子数据、Docker 一键启动、单元测试 + Playwright E2E（Mock，¥0）。

### V1（本轮不做）

真实 OSS/COS 联调验收、邮件模板可视化编辑、提醒规则批量导入、家庭账单导出 CSV、多家庭切换 UI 优化。

### V2（本轮不做）

原生 App、微信小程序、AI 症状诊断、商城支付、社交动态、i18n、SaaS 计费、短信/电话通道。

---

## 目录结构（冻结）

```
backend/                 # Go 服务（handler → service → repo，禁止反向）
frontend-user/           # Vue 3 用户端（本轮唯一前端）
frontend-admin/          # 占位：需求未列管理后台，禁止实现
frontend-mp/             # 占位：需求明确不做小程序
tests/                   # Playwright E2E + API smoke
docs/
docker-compose.yml
```

依赖方向：`handler → service → repo`；`reminder` / `storage` / `notifier` / `ws` 只被 service 调用。

---

## 任务清单

### Phase 1 · Architect

- [x] git init + .gitignore
- [x] 本文件（MVP/V1/V2 + Phase Order）
- [x] docker-compose.yml 随机端口 15231–15236
- [x] 严格目录骨架

### Phase 3 · Logic（先于 UI）

- [x] T3-1 配置 / 时区 / Logger / 错误码 / JWT / RBAC 中间件
- [x] T3-2 SQL migration + 种子数据（2 宠物 × 12 月）
- [x] T3-3 Auth / Family / Pet CRUD
- [x] T3-4 Checkin + WebSocket 房间隔离
- [x] T3-5 Health events + 动态重算挂钩
- [x] T3-6 Reminder 引擎 + Notifier（SMTP/WeCom/Webhook）
- [x] T3-7 Expense / Weight 聚合 API
- [x] T3-8 Storage 抽象（local/oss/cos）+ 鉴权下载
- [x] T3-9 Dockerfile + 核心单测
- [x] T3-10 docs/API.md + docs/.meta/api_contracts.md

### Phase 2 · UI

- [x] T2-1 docs/DesignSpec.md（暖木 / 奶油 / 苔绿宠物管家美学）
- [x] T2-2 登录注册 / 家庭邀请
- [x] T2-3 宠物主页卡片（年龄 + 今日打卡）
- [x] T2-4 健康时间轴
- [x] T2-5 Echarts 体重曲线 + 分类开销
- [x] T2-6 相册 / PDF / 通知中心 / 提醒规则
- [x] T2-7 四态反馈 + 375/768/1280 响应式

### Phase 4 · QA

- [x] T4-1 单元测试覆盖核心包 ≥ 70%
- [x] T4-2 Playwright ≥ 5 条关键路径（Mock，¥0）
- [x] T4-3 异步可靠性：重启恢复 / 幂等 / 失败分类
- [x] T4-4 docs/QA_Record.md

### Phase 5 · Auditor

- [x] T5-1 对照 audit-rules.md 出 AuditReport
- [x] T5-2 Knowledge Harvest

---

## 开发端口（随机，Deploy 阶段再改为 8081+）

| 服务 | 宿主端口 | 容器端口 |
|---|---|---|
| frontend-user (nginx) | 27151 | 80 |
| backend API + WS | 27152 | 8080 |
| postgres | 27153 | 5432 |
| redis | 27154 | 6379 |
| mailhog SMTP | 27155 | 1025 |
| mailhog UI | 27156 | 8025 |

---

## 变更记录

| 版本 | 时间 | 变更 |
|---|---|---|
| v1.0 | 2026-08-23 16:07 | Phase 1 冻结。Logic-First。MVP 覆盖 F1–F8。 |
