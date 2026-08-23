# GoPuppy API

Base: `http://localhost:27152/api/v1`（经前端反代则为 `http://localhost:27151/api/v1`）  
统一响应：`{ "code", "message", "request_id", "data" }`  
时间展示约定：用户可见 `yyyy-MM-dd HH:mm:ss`（GMT+8）。

## 错误码

| HTTP | code | 含义 |
|---|---|---|
| 401 | UNAUTHORIZED | 未登录 / Token 无效 |
| 404 | NOT_FOUND | 资源不存在或跨家庭（刻意不返回 403） |
| 409 | CONFLICT | 邮箱占用 / 已是成员 |
| 413 | TOO_LARGE | 上传超限 |
| 422 | VALIDATION | 字段/类型/边界不合法 |
| 500 | INTERNAL | 内部错误 |

## Auth

### POST /auth/register

请求
```json
{ "email": "new@gopuppy.test", "password": "Puppy123!", "nickname": "新成员" }
```
响应 `201`
```json
{ "code": "CREATED", "data": { "user": { "id": "...", "email": "new@gopuppy.test", "nickname": "新成员" }, "tokens": { "access_token": "...", "refresh_token": "...", "expires_at": "..." } } }
```

### POST /auth/login

请求 `{ "email": "dad@gopuppy.test", "password": "Puppy123!" }`  
响应同登录成功，`code=OK`。

### POST /auth/refresh

请求 `{ "refresh_token": "..." }` → `{ "access_token", "refresh_token", "expires_at" }`

### GET /me

Header `Authorization: Bearer <access>` → 当前用户。

## Family

### GET /families
### POST /families `{ "name": "林家小院" }`
### GET /families/{familyID}/members
### POST /families/{familyID}/invites `{ "role": "CAREGIVER" }` → `{ "code": "AB12CD", "expires_at": "..." }`
### POST /families/join `{ "code": "AB12CD" }`
### DELETE /families/{familyID}/members/{userID}

## Pets

### GET /families/{familyID}/pets
响应含 `age: { years, months, days, total_days }`（后端按 GMT+8 计算）。

### POST /families/{familyID}/pets
```json
{ "name": "奶油", "species": "CAT", "breed": "英短", "gender": "FEMALE", "birthday": "2023-03-15", "neutered": true }
```

### GET /pets/{petID}
### PATCH /pets/{petID}
### DELETE /pets/{petID}  （OWNER 软归档）

## Checkins + WebSocket

### GET /pets/{petID}/checkins/today
### POST /pets/{petID}/checkins
```json
{ "type": "FEED", "slot": "MORNING", "done": true }
```

WebSocket：`ws://localhost:27151/ws?token=<access>&family_id=<uuid>`  
无 token → 401。房间键 = family_id。消息：`{ "type": "checkin.updated", "pet_id", "payload" }`。

## Health events

### GET /pets/{petID}/events?category=VACCINE&year=2025
### POST /pets/{petID}/events
```json
{ "category": "VACCINE", "title": "狂犬疫苗", "occurred_at": "2026-08-23 15:00:00", "clinic": "萌宠医院" }
```
补录 VACCINE/DEWORM/CHECKUP/MEDICATION 会立刻重算对应提醒规则的 `next_due_at`。

## Reminders

### GET /pets/{petID}/reminders
### POST /pets/{petID}/reminders
```json
{ "kind": "DEWORM", "title": "体内驱虫", "cycle_days": 90, "last_done_at": "2026-07-20 09:00:00", "advance_days": 3, "channels": ["EMAIL", "WEBHOOK"] }
```
### GET /families/{familyID}/notifications
### POST /notifications/{logID}/replay
### POST /admin/reminder-scan  （开发用手动触发当日扫描）

## Finance

### GET /pets/{petID}/finance  后端按月聚合 12 个月
### POST /pets/{petID}/weights `{ "weight_kg": 4.8, "measured_at": "2026-08-23 10:00:00" }`
### POST /pets/{petID}/expenses `{ "category": "FOOD", "amount_cents": 16800, "spent_at": "2026-08-23 10:00:00" }`

金额单位：整数分。

## Media

### GET /pets/{petID}/media?kind=PHOTO
### POST /pets/{petID}/media  multipart `file` + `kind`
白名单：JPEG/PNG/WEBP/PDF（magic bytes），单文件 ≤ 20MB。
### GET /media/{mediaID}/file  需登录，按宠物家庭鉴权。

## Health

### GET /healthz
### GET /readyz  探测 DB + Redis + 存储驱动名
