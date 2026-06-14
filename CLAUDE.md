# kun-galgame-forum（kungal）— AI 代理项目指南

视觉小说 / galgame **论坛**。`apps/api` = Go Fiber v3 + GORM + Postgres，`apps/web` = Nuxt 4。
本仓是 **kun-galgame-infra（OAuth / 身份 / 契约中枢）的下游**之一（另一个是 kun-galgame-patch / moyu）。

## 跨服务契约（不可违反 — 由 kun-galgame-infra 拥有）

权威契约文档以**只读镜像**同步在 `docs/{oauth,image_service,galgame_wiki}/`（文件头有 GENERATED banner）。
**改契约请去 infra 源头改，别动这里的副本**；副本由 kungal-docs 的 `pnpm docs:sync` 重新生成。核心不变量：

- **身份（C1/C2）**：`user.id` 在本库、OAuth、另一下游中是**同一个整数**——永不重新编号用户；本地表用 `*_user_id` 对齐 OAuth `users.id`。OAuth 拥有身份并签发 JWT，本服务只**验签、从不签发**（见 `internal/middleware/auth.go`）。
- **用户资料（C6）**：**不持久化到本地 user 表、不当真源**（短 TTL 内存缓存可以——`pkg/userclient` 已内置约 10min TTL）；按 id 列表走 `GET /users/batch`（OAuth Client Basic Auth，≤100 个 id，**不返回** email / moemoepoint / created_at）。@提及补全用 `GET /users/search`（**勿缓存**）；当前用户用 `/oauth/userinfo`。OAuth 不发 SDK，自己实现薄客户端。
- **萌萌点 moemoepoint（C3）**：每用户单一余额，**单源在 OAuth**；本地 `users.moemoepoint` 是缓存视图，不可当真源。发放/扣除走 s2s API，幂等键 = `<app>:<event>:<ref>`（如 `kungal:liked:topic_1207`）。下游可用 reason：`content_approved` / `content_removed` / `daily_checkin` / `liked`；**OAuth 保留、s2s 禁用**：`admin_grant` / `admin_deduct` / `migration` / `register_gift`。推送在 `internal/moemoepoint/pusher.go`。s2s 端点**已实现**（`POST/GET /users/:id/moemoepoint`，`Adjust` 幂等；见 infra `internal/platform/auth/handler/moemoepoint_handler.go` 与 `cmd/oauth/main.go`）。
- **图片（C4）**：内容寻址图床在 OAuth（头像 / 共享图**不走本地 S3**）。URL = `{base}/{aa}/{bb}/{hash}[_variant].webp`（两级十六进制分片）；传递 `*_image_hash` 字段，用 image client 解析。
- **Wiki 消息（C5）**：galgame-wiki 拥有 `galgame_message`；消费 `GET /galgame/messages/mine`（通知中心）与 `/galgame/messages/feed`（cron）。无 target 的消息只进 admin 队列。

完整细节见 `docs/oauth/`、`docs/image_service/`、`docs/galgame_wiki/`。

## 本仓要点

- 鉴权在 `internal/middleware/auth.go`：是 **BFF 不透明会话**（`kungal_session` cookie + Redis 存 OAuth token），不是纯 JWT 验签。会话 **90 天滑动续期**（活跃用户不再每周掉线）——模型与 2026-06 修复见 `docs/proj/session-lifetime.md`。
- 修改 `docs/{oauth,image_service,galgame_wiki}/` 下任何文件都是**错的**——那是 infra 的镜像；要改去 infra 改、再 `docs:sync`。
- **改了数据库 schema 必须提醒迁移**：本仓 schema 走 `apps/api/migrations/NNN_*.up.sql`（幂等，`IF NOT EXISTS`）+ 自带 migrate runner，**启动时不跑 AutoMigrate**。只要加了迁移文件 / 改了表结构，**任务结束时必须明确告诉用户：是否需要在生产跑迁移、跑哪个命令**。漏跑 → 线上代码读到不存在的列 → 静默故障（参考 2026-06 infra 萌萌点发放故障：缺一列导致全站 ~29h 拿不到萌萌点）。
