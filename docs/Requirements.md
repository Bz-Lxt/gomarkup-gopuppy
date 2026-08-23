# GoPuppy · 需求规格说明书（SSOT）

> **权威声明**：本文件定义 **WHAT**（做什么）。`docs/Roadmap.md` 定义 **WHEN**（何时做）。
> 任何代码实现不得超出本文件范围（Redline 4 · NO SCOPE DRIFT）。

| 项 | 值 |
|---|---|
| 项目代号 | GoPuppy |
| 中文名 | 宠物全能助手 · 一站式宠物健康档案与全生命周期智能管家 |
| 需求冻结时间 | 2026-08-23 15:44 (GMT+8) |
| 原始 Prompt | `docs/.meta/original_prompt.md` |
| PM 结论 | ✅ **ACCEPT**（Tier 2 规模，强制分期 Roadmap） |

---

## 1. PM 准入评估记录（Gatekeeper Audit Trail）

### 1.1 废除判据逐项核验

| # | 判据 | 结论 | 依据 |
|---|---|---|---|
| 1 | 不完整 / 模糊 | ✅ 通过 | 主题明确（养宠家庭管理系统），六大功能域描述具体，无缺失附件依赖 |
| 2 | Windows 独占 | ✅ 通过 | Go + Web 前端 + Docker，全平台 |
| 3 | 规模评估（分级） | ✅ **ACCEPT + 强制分期** | 见 §1.2 |
| 4 | 外部依赖（智能判定） | ✅ **ACCEPT + Mock Provider** | 见 §1.3 |
| 5 | 专业 / 小众付费软件 | ✅ 通过 | Go / Vue / PostgreSQL / Redis / Docker 全部开源免费 |

### 1.2 规模评估（Tier 2：10,000 – 40,000 LoC）

| 模块 | 文件数估算 | LoC 估算 |
|---|---|---|
| `backend`（Go） | 28 – 36 | 5,000 – 7,500（用户指定） |
| `frontend-user`（Vue 3） | 38 – 48 | 4,000 – 5,500 |
| 基础设施（Docker / SQL / CI） | 8 – 12 | 500 – 800 |
| 测试（Go unit + Playwright E2E） | 10 – 14 | 800 – 1,200 |
| **合计** | **84 – 110** | **≈ 10,300 – 15,000** |

**判定**：落入 `10,000 – 40,000 LoC` 区间 → **ACCEPT**。
**强制约束**：Phase 1 必须产出带明确 **MVP / V1 / V2 边界** 的 `docs/Roadmap.md`，未产出前禁止写任何业务代码。

### 1.3 外部依赖智能判定（Scenario A / B）

| 外部依赖 | 场景 | 判定 | Mock 策略 |
|---|---|---|---|
| SMTP 邮件推送 | **A · 可模拟** | ACCEPT | `MailHog` 容器接收真实 SMTP，Web UI 可视化验收（非伪造） |
| 企业微信群机器人 | **A · 可模拟** | ACCEPT | 本地 Mock Webhook 端点，落库 + 日志可查 |
| 通用 Webhook | **A · 可模拟** | ACCEPT | 同上，共用投递抽象层 |
| 阿里云 OSS / 腾讯云 COS | **A · 可模拟** | ACCEPT | `Storage` 接口 + `local` 驱动为默认实现，OSS/COS 驱动同步落地但默认关闭 |

**无 Scenario B 依赖**：本系统不需要实时事实性数据（股价 / 赛事 / 路况），Mock 可完整替代。

**Mock 合法性承诺（Redline 4）**：以上每一项都必须同时满足
（a）真实实现路径已存在且已接线；
（b）Mock/Real 切换开关在 `README.md` §7 明确文档化。
否则视为**造假**，触发 STOP。

---

## 2. 兼容性与逻辑冲突核验（Step 2）

### 2.1 交付标准核验（Redline 1）

- **微信小程序例外**：❌ 不适用。企业微信机器人仅为**通知出口**，不涉及小程序运行时 → **Docker 检查照常执行**。
- **Docker 可行性**：✅ Go 编译为静态二进制 + Vue 构建静态资源，天然容器化。
- **跨平台核验**（已实测 `docker pull --platform linux/arm64`）：

| 镜像 | ARM64 |
|---|---|
| `golang:1.23-alpine` | ✅ |
| `postgres:16-alpine` | ✅ |
| `node:20-alpine` | ✅ |
| `nginx:1.27-alpine` | ✅ |
| `redis:7-alpine` | ✅ |

- **localhost 可访问服务**：Web 前端 + REST API + WebSocket + MailHog UI，全部经 `localhost` 暴露。

### 2.2 矛盾与待决项裁决（Contradiction Detection）

| # | 原文表述 | 性质 | PM 裁决 |
|---|---|---|---|
| C1 | "Vue 3 / React + Tailwind" | 二选一未定 | **裁定 Vue 3 + TypeScript + Tailwind CSS + Pinia**。理由：用户表述中 Vue 3 居首；Echarts 官方 Vue 集成成熟；单一栈避免 Phase 2 分裂。 |
| C2 | "妈妈的**手机端**通过 WebSocket 实时看到" | 隐含原生 App | **裁定不做原生 App**。交付**移动优先响应式 Web**（≥ 375px 断点完整可用），符合 Redline 1 的 localhost 可访问要求。原生 App 属 V2 范围外。 |
| C3 | "多角色共同养育（RBAC）"但未定义角色 | 定义缺失 | **PM 补全**为 3 角色模型，见 §3.1。 |
| C4 | "Echarts 渲染**近一年**体重变化" | 新系统无历史数据 → 空图表 | **强制要求种子数据**：至少 2 只宠物 × 12 个月体重记录 × 分类开销记录，随 migration 自动注入。 |
| C5 | "年龄精确到天" + "当天推送提醒" | 时区陷阱 | **强制 GMT+8**：Go 侧统一 `Asia/Shanghai` 时区常量，容器 `TZ=Asia/Shanghai`，DB 存 `TIMESTAMPTZ`。禁止裸用 `time.Now().UTC()` 做日界判断（UTC 会导致日期偏移 8 小时，08:00 前的提醒全部错日）。 |
| C6 | "异常症状记录"与"健康时间轴"关系未明 | 归属不明 | **裁定合并**：症状记录为 `health_events` 的一个 `category=SYMPTOM`，统一进时间轴，避免双表冗余。 |
| C7 | "病历单、体检 PDF 上传"未限定类型 / 大小 | 安全缺口 | **PM 补全**安全基线，见 §5.3。 |

### 2.3 给 Chief Architect 的 Phase Order 输入（Phase 1 决策项）

本项目 UI 的两个核心组件——**健康时间轴（Timeline）** 与 **Echarts 体重/开销可视化**——其组件结构直接派生于数据模型（事件分类枚举、开销分类枚举、时间序列聚合粒度）。按 SOP v13 §4 Phase 1.5 判据，**建议采用 Logic-First（交换 Phase 2 / Phase 3）**，或至少在 Phase 2 开工前冻结 §6 数据模型与枚举。最终决策权归 Chief Architect，须在 `docs/Roadmap.md` 记录一句话理由。

---

## 3. 功能需求

### 3.1 F1 · 账户与家庭 RBAC

**家庭模型**：`User` ↔ `Family` 多对多（带角色），`Pet` → `Family` 多对一。爸爸与妈妈加入同一 `Family` 即共同拥有该家庭全部宠物。

| 角色 | 代号 | 权限 |
|---|---|---|
| 家庭主人 | `OWNER` | 全部权限 + 成员邀请/移除/改角色 + 删除宠物 + 删除家庭 |
| 共同养育者 | `CAREGIVER` | 宠物档案读写、打卡、记账、上传媒体、管理提醒规则；**不可**管理成员或删除宠物 |
| 只读访客 | `VIEWER` | 仅读（供亲友 / 寄养方查看）；不可写入任何数据 |

**需求项**
- F1-1 邮箱 + 密码注册登录，密码 `bcrypt` 哈希（cost ≥ 10），禁止明文/可逆存储。
- F1-2 JWT 认证（Access Token 2h + Refresh Token 7d），Token 内携带 `user_id`，**不携带角色**（角色每请求按 `family_id` 实时查库，防越权提权）。
- F1-3 邀请码加入家庭：`OWNER` 生成 6 位邀请码（含有效期 24h），成员凭码加入并被赋予指定角色。
- F1-4 **权限中间件强制**：所有涉及 `pet_id` / `family_id` 的端点必须校验"当前用户 ∈ 该家庭 且 角色满足操作最低要求"。跨家庭访问返回 `404`（不返回 `403`，避免资源存在性泄露）。
- F1-5 审计字段：所有写操作记录 `created_by` / `updated_by`，用于"谁喂的食"归属展示。

### 3.2 F2 · 宠物档案

- F2-1 字段：名称、物种（猫/狗/其他）、品种、性别、**出生日期（必填，年龄计算基础）**、头像、是否绝育、芯片号、体重目标区间、备注。
- F2-2 **年龄精确到天**：后端按 GMT+8 计算并返回 `{years, months, days, total_days}` 结构化字段，前端不做日期算术（避免客户端时区污染）。
- F2-3 归档而非硬删除：`archived_at` 软删除，保留历史档案（宠物"全生命周期"语义要求）。
- F2-4 一个家庭支持 ≥ 10 只宠物。

### 3.3 F3 · 每日打卡 + WebSocket 实时同步

- F3-1 打卡类型：`FEED`（喂食）、`MEDICINE`（吃药）；时段 `MORNING` / `NOON` / `NIGHT`。
- F3-2 **幂等唯一约束**：`UNIQUE(pet_id, date, type, slot)`。重复勾选不产生脏数据，取消勾选为软撤销。
- F3-3 宠物主页卡片展示"今日已喂食/已吃药"快捷勾选，含**操作人昵称与时间**（"妈妈 08:12 已喂食"）。
- F3-4 **WebSocket 实时广播**：爸爸打卡后，同家庭在线成员在 **≤ 1 秒**内收到状态变更，无需刷新。
  - 连接鉴权：`ws://.../ws?token=<JWT>`，握手阶段校验，失败立即关闭。
  - 房间隔离：按 `family_id` 分房，**严禁跨家庭消息泄露**。
  - 断线重连：前端指数退避重连（1s → 2s → 4s → 上限 30s），重连后拉取一次全量今日状态补偿。
  - 心跳：服务端 30s ping，客户端未响应 60s 判定离线并回收。
- F3-5 日界翻转按 GMT+8 的 `00:00`。

### 3.4 F4 · 健康时间轴

- F4-1 事件分类：`VACCINE`（疫苗）、`DEWORM`（驱虫）、`SURGERY`（手术/绝育）、`CHECKUP`（体检）、`SYMPTOM`（异常症状）、`MEDICATION`（用药）、`OTHER`。
- F4-2 事件字段：分类、标题、详细描述、发生时间、就诊机构、费用（可选，联动 F6 记账）、附件（联动 F7）、下次到期时间（可选，联动 F5）。
- F4-3 前端以**精美垂直时间线**按时间倒序渲染，分类配色 + 图标区分，支持按分类筛选与年份跳转。
- F4-4 `SYMPTOM` 事件支持严重度标记（`MILD` / `MODERATE` / `SEVERE`）与"是否已就医"布尔位。

### 3.5 F5 · 智能动态提醒引擎（后端核心）

- F5-1 **提醒规则**：`kind`（VACCINE / DEWORM / MEDICINE / CHECKUP）+ `cycle_days`（周期天数）+ `last_done_at` → 引擎计算 `next_due_at`。
- F5-2 **动态重算**：用户补录一条 `VACCINE` 健康事件后，对应规则的 `next_due_at` 必须自动前推重算（这是"动态"的核心语义，非静态定时）。
- F5-3 **定时扫描**：每日 GMT+8 `08:00` 由 Go 定时任务扫描当日到期规则，**调度触发误差 ≤ 60 秒**。
- F5-4 **多通道投递**：`EMAIL` / `WECOM_BOT` / `WEBHOOK`，规则级可多选。通道通过统一 `Notifier` 接口抽象。
- F5-5 **投递幂等**：`UNIQUE(rule_id, due_date, channel)` 幂等键。服务重启、任务重跑、多实例并发均不得重复推送。
- F5-6 **窄重试（Narrow Retry）**：仅对**瞬时错误**重试（网络超时 / `5xx` / `429`），最多 3 次指数退避（2s / 8s / 32s）。`401` / `403` / `422` 等**认证与校验类错误一律不重试**并落库标记 `PERMANENT_FAILURE`。
- F5-7 **可靠性**：任务执行状态落库（`PENDING` / `SENT` / `FAILED` / `PERMANENT_FAILURE`），容器重启后未完成任务可恢复；提供手动重放接口。
- F5-8 **提前提醒**：支持到期前 N 天（默认 3 天）预警，与当日提醒各自独立幂等。
- F5-9 提前提醒与到期提醒的投递记录对用户可见（通知中心页面）。

### 3.6 F6 · 记账与体重曲线

- F6-1 开销分类：`FOOD`（猫粮/狗粮）、`MEDICAL`（医疗）、`TOY`（玩具）、`GROOMING`（美容）、`INSURANCE`（保险）、`OTHER`。
- F6-2 金额以**整数分（cents）**存储，禁止浮点，避免累加精度丢失。
- F6-3 体重记录：`weight_kg`（`NUMERIC(5,2)`）+ 测量时间 + 备注。
- F6-4 **Echarts 可视化**：
  - 体重折线图（近 12 个月，含目标区间参考带，异常波动 > ±10% 标红点）。
  - 开销分类饼图 + 月度堆叠柱状图（近 12 个月）。
  - 聚合在**后端**完成（返回已按月分桶的序列），前端不做重计算。
- F6-5 统计概览：本月总开销、单只宠物年度累计、分类占比 Top3。

### 3.7 F7 · 多媒介存储隔离

- F7-1 **存储抽象**：`Storage` 接口（`Put` / `Get` / `Delete` / `SignedURL`），驱动 `local`（默认）、`oss`、`cos`，通过环境变量 `STORAGE_DRIVER` **动态切换，无需改代码**。
- F7-2 **按宠物 ID 目录隔离**：`{root}/pets/{pet_id}/{kind}/{yyyy-MM}/{sha256}.{ext}`。
- F7-3 媒体类型：`PHOTO`（云端相册）、`MEDICAL_RECORD`（病历单）、`REPORT_PDF`（体检报告）。
- F7-4 **内容寻址去重**：以文件 SHA-256 作为对象键，同一家庭内重复上传复用已有对象（省空间且天然幂等）。
- F7-5 **访问控制**：文件下载必须经后端鉴权代理或短时效签名 URL（默认 15 分钟），**严禁静态目录直接暴露**（否则任意人可遍历他人宠物病历）。
- F7-6 相册前端瀑布流 + 灯箱预览；PDF 提供下载与浏览器内预览。

### 3.8 F8 · 系统与运维

- F8-1 **统一 Logger**（对应全局记忆 `[Logging]`）：结构化日志（`slog`），level 可配（`debug`/`info`/`warn`/`error`），生产环境自动屏蔽 debug。**禁止散落 `fmt.Println`**。
- F8-2 **统一错误响应**：`{code, message, request_id}`，全局 `request_id` 贯穿日志便于追踪。
- F8-3 `/healthz`（存活）+ `/readyz`（依赖就绪：DB / Redis / 存储）。
- F8-4 **API 文档**（对应全局记忆 `[Documentation]`）：`docs/API.md` 必须含每个端点的**请求/响应示例 + 参数类型说明 + 错误码表**，仅有清单不算达标。
- F8-5 **外部数据校验**（对应全局记忆 `[Robustness]`）：所有入参（含 JSON 反序列化、文件上传、Webhook 回调）必须做字段存在性、类型、边界值校验，不得仅依赖调用方。
- F8-6 数据库 migration 版本化管理，支持前滚；含种子数据（对应 C4）。

---

## 4. 非功能需求

| 维度 | 要求 |
|---|---|
| 交付方式 | `docker compose up --build -d` 一键启动，零手工配置 |
| 架构 | ARM64 + AMD64 双架构；多阶段构建 |
| 时区 | 全链路 GMT+8（`TZ=Asia/Shanghai`），DB 使用 `TIMESTAMPTZ` |
| 前端美学 | Redline 2「Dribbble 标准」：Tailwind 设计系统，统一间距/字阶/圆角/阴影，含 loading / empty / error / success 四态反馈 |
| 响应式 | 移动优先，375px / 768px / 1280px 三断点完整可用 |
| 安全 | 见 §5 |
| 可观测 | 结构化日志 + request_id + 通知投递审计表 |
| 测试 | 见 §5.4 |

---

## 5. 验收基线（可测量 · Measurable Acceptance Criteria）

> 本节全部为**数字化门槛**，非散文描述。Phase 4 QA 与 Phase 5 Auditor 按此逐条核验。

### 5.1 性能

| 指标 | 门槛 |
|---|---|
| REST API P95 延迟（本地 Docker，Mock 模式） | **< 200 ms** |
| 首屏 LCP（本地，构建产物） | **< 2.5 s** |
| WebSocket 状态同步端到端延迟（同家庭） | **≤ 1 s** |
| Echarts 12 个月数据渲染 | **< 500 ms** |
| 提醒任务调度触发误差 | **≤ 60 s** |

### 5.2 正确性

| 指标 | 门槛 |
|---|---|
| 年龄计算精度 | 误差 **0 天**（GMT+8 基准，含闰年 2/29 边界） |
| 打卡幂等 | 重复提交 N 次，DB 记录数恒为 **1** |
| 通知幂等 | 同一规则同一天同一通道，实际投递 **恰好 1 次**（任务重跑后仍为 1） |
| 提醒动态重算 | 补录健康事件后 `next_due_at` **立即更新**，误差 0 天 |
| 跨家庭越权 | 任意跨家庭读写尝试，**100% 返回 404** |
| 金额精度 | 1000 笔随机开销累加，与十进制精确值误差 **0 分** |

### 5.3 安全

| 项 | 门槛 |
|---|---|
| 密码存储 | `bcrypt` cost ≥ 10；数据库中**不存在**任何明文密码 |
| 上传大小 | 单文件 **≤ 20 MB**，请求体总量 ≤ 50 MB |
| 上传类型白名单 | `image/jpeg`、`image/png`、`image/webp`、`application/pdf`，**以 magic bytes 嗅探判定，不信任扩展名与 Content-Type** |
| 文件名安全 | 拒绝路径穿越（`../`、绝对路径、NUL 字节）；落盘名由 SHA-256 生成，不使用用户原名 |
| 文件访问 | 无有效凭证访问他人宠物媒体，**100% 拒绝** |
| SQL 注入 | 全部参数化查询，0 处字符串拼接 SQL |
| 密钥管理 | 无任何密钥/口令硬编码入库；全部经环境变量注入，`.env.example` 提供占位 |
| WebSocket 鉴权 | 无 token 或无效 token 握手，**100% 拒绝** |
| CORS | 显式白名单，禁止 `*` 配合凭证 |

### 5.4 测试（对应全局记忆 `[Testing]`）

| 项 | 门槛 |
|---|---|
| 后端单元测试 | 核心包（domain / service / reminder engine / storage）语句覆盖率 **≥ 70%**；CRUD + 提醒引擎必测 |
| 提醒引擎专项 | 必须覆盖：周期计算、动态重算、幂等、窄重试分类（瞬时 vs 永久）、时区日界 |
| E2E（Playwright） | **≥ 5 条**关键路径：①注册→建家庭→建宠物 ②今日喂食打卡 ③双端 WebSocket 同步 ④补录疫苗事件→时间轴与下次到期更新 ⑤上传照片/PDF→相册与鉴权下载 |
| **测试成本** | E2E 与 smoke 全程 **Mock / 离线模式**，禁止调用计费 API，单轮 QA 预期支出 **¥0** |
| 异步可靠性 | 必测：容器重启后未完成提醒任务恢复；重跑任务幂等；失败分类正确 |

### 5.5 交付完整性

- `docker compose up --build -d` 后 **≤ 120 秒**内所有服务 healthy。
- `README.md` 含 7 个强制章节（§1 如何启动 … §7 API 模拟与切换指南）。
- 种子数据可见：登录测试账号后，时间轴 / 体重曲线 / 开销图表**均非空**。

---

## 6. 数据模型草案（Phase 1 可细化，枚举值冻结）

```
users              (id, email✱, password_hash, nickname, avatar_url, created_at, updated_at)
families           (id, name, owner_id→users, created_at)
family_invites     (id, family_id, code✱, role, expires_at, used_by, used_at)
family_members     (family_id, user_id, role, joined_at)  PK(family_id,user_id)
pets               (id, family_id, name, species, breed, gender, birthday, avatar_key,
                    neutered, chip_no, weight_min, weight_max, note, archived_at, created_at)
daily_checkins     (id, pet_id, checkin_date, type, slot, done_by→users, done_at, revoked_at)
                    UNIQUE(pet_id, checkin_date, type, slot)
health_events      (id, pet_id, category, title, description, occurred_at, clinic,
                    severity, treated, amount_cents, created_by, created_at)
event_attachments  (event_id, media_id)
reminder_rules     (id, pet_id, kind, title, cycle_days, last_done_at, next_due_at,
                    advance_days, channels[], enabled, created_at)
notification_logs  (id, rule_id, pet_id, due_date, channel, status, attempt, error,
                    scheduled_at, sent_at)  UNIQUE(rule_id, due_date, channel, kind_advance)
weight_records     (id, pet_id, weight_kg, measured_at, note, created_by)
expenses           (id, pet_id, category, amount_cents, spent_at, note, created_by)
media_files        (id, family_id, pet_id, kind, storage_driver, object_key, filename,
                    mime, size_bytes, sha256✱, uploaded_by, created_at)
```

**冻结枚举**（前后端共用，禁止各自定义）
- `Role`: `OWNER` / `CAREGIVER` / `VIEWER`
- `CheckinType`: `FEED` / `MEDICINE`；`Slot`: `MORNING` / `NOON` / `NIGHT`
- `EventCategory`: `VACCINE` / `DEWORM` / `SURGERY` / `CHECKUP` / `SYMPTOM` / `MEDICATION` / `OTHER`
- `Severity`: `MILD` / `MODERATE` / `SEVERE`
- `ReminderKind`: `VACCINE` / `DEWORM` / `MEDICINE` / `CHECKUP`
- `Channel`: `EMAIL` / `WECOM_BOT` / `WEBHOOK`
- `ExpenseCategory`: `FOOD` / `MEDICAL` / `TOY` / `GROOMING` / `INSURANCE` / `OTHER`
- `MediaKind`: `PHOTO` / `MEDICAL_RECORD` / `REPORT_PDF`
- `StorageDriver`: `local` / `oss` / `cos`
- `NotifyStatus`: `PENDING` / `SENT` / `FAILED` / `PERMANENT_FAILURE`

---

## 7. 风险登记（Risk Register）

| # | 风险 | 影响 | 缓解措施 |
|---|---|---|---|
| R1 | UTC / GMT+8 时区偏移导致"当天提醒"错日、年龄差 1 天 | 高 | 统一时区常量 + 容器 `TZ` + 时区专项单测（含 00:00 与 23:59 边界） |
| R2 | 定时任务重启或多实例导致重复推送 | 高 | 幂等唯一键 + 状态落库 + 重启恢复测试 |
| R3 | 文件静态目录直接暴露导致病历泄露 | 高 | 强制鉴权代理 / 签名 URL，Auditor 专项核验 |
| R4 | WebSocket 房间隔离失效导致跨家庭数据泄露 | 高 | 房间键 = `family_id`，E2E 双账号跨家庭负向测试 |
| R5 | 无 OSS/COS 真实凭证，驱动实现无法验证 | 中 | Contract Gate 标记 `UNVERIFIED`；`local` 驱动为默认路径并完整测试；README §7 说明切换方式 |
| R6 | Tier 2 规模导致上下文超限、后期返工 | 中 | 强制 MVP/V1/V2 分期；枚举与数据模型 Phase 1 冻结 |
| R7 | 前端图表空数据观感崩塌 | 中 | 强制种子数据（C4）+ empty state 设计 |
| R8 | 25+ Go 文件下分层混乱 | 中 | Phase 1 冻结目录结构与依赖方向（handler → service → repo，禁止反向） |

---

## 8. 范围边界（明确不做 · Anti Scope Drift）

- ❌ 原生 iOS / Android App（响应式 Web 覆盖，见 C2）
- ❌ 微信小程序端
- ❌ 在线宠物商城 / 支付结算
- ❌ 宠物社交社区 / 动态流
- ❌ AI 症状诊断与用药建议（医疗合规风险，且原始 Prompt 未要求）
- ❌ 多语言 i18n（默认中文单语）
- ❌ 多租户 SaaS 计费体系
- ❌ 真实短信 / 电话通知通道

---

## 9. 变更记录

| 版本 | 时间 (GMT+8) | 变更 |
|---|---|---|
| v1.0 | 2026-08-23 15:44 | PM Agent 首次冻结。准入 ACCEPT（Tier 2）；裁决 C1–C7；补全 RBAC 角色模型、安全基线、可测量验收基线。 |
