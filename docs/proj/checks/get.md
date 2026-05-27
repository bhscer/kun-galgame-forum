# GET API 字段对齐检查

> 目的: 记录全部 GET 端点 (查询类) 以及 FE↔BE 字段对齐审计状态。
>
> 路由源: `apps/api/internal/app/router.go`
>
> 配套文档: [post.md](./post.md) · [put.md](./put.md) · [delete.md](./delete.md)
>
> **本文件目前是 inventory 阶段** — 仅列出全部 GET 端点供后续逐项审计，状态列暂为 ⏳ (待审计)。

## 图例

- ✅ 已审计，FE/BE 对齐无问题
- 🔧 已审计，**发现错位并修复**
- ⏭️ 已审计，设计上有意保持当前行为 (详见备注)
- ⏳ 待审计 (本轮 inventory 阶段默认)

## 鉴权图例 (中间件分组)

| 标记 | 中间件 | 含义 |
|---|---|---|
| 🌐 | (无) | 完全公开，任何来源可访问 |
| 🔐 | `optAuth` | 鉴权可选；带 token 时附加自己的 pending/隐藏内容，匿名时只看公共部分 |
| 🔒 | `userAuth` / `authed` | 必须登录 (`Auth` 中间件)，401 if missing |
| 🛡️ | `galgameAdmin` (`role >= 2`) | 仅 admin/moderator |
| ⚙️ | `admin` (`role >= 3`) | 仅 admin |

## 后端 handler 类型

- **kungal-native** — 直接调用 kungal Go 业务 handler，可能附加 wiki 调用 + 本地数据装配
- **proxy** — `GalgameWikiHandler.ProxyGet`，透传到 wiki service 的同名/经 `path_mapper` 改写的端点
- **wiki+local** — handler 同时调 wiki API 并合并本地数据 (如 `GetGalgameLinks`, `GetGalgameHistory`)

## 统计

- 全部 GET 端点 (展开 for-loop 后): **99**
  - 静态注册行: 93 (原 95，K-PR 第一轮审计删了 `/galgame/check` + `/galgame/:gid/contributors`)
  - 实体 revision 循环 (`tag` / `official` / `engine` × 2 routes): 6
- 已审计：**§0 + §1 + §2 + §3 + §4 + §5 完成**（第一组），**§6 + §7 + §8 + §9 + §10 + §11 完成**（第二组）
- 已修复 (本轮 K-PR 审计):
  - 第一组 (§0–§5)：
    - `/admin/setting/register` FE 状态反转
    - `/ranking/{galgame,topic,user}` FE 泛型 (UI metadata → API DTO)；`/ranking/user` BE 补 `SortField`；`/ranking/galgame` BE 补 `effective_banner_hash/_url`
    - `/section`, `/category` FE 缺失类型 (`SectionTopicList` 补上，`CategorySection` → `CategorySectionStats`)
    - `/search` reply/comment/user FE 字段守护 (optional + v-if)
    - `/topic`, `/resource` `isPollTopic` BE hardcode false → 批量 `FindTopicIDsWithPoll`
    - `/topic/:tid/poll/topic` FE 泛型 (单 → 数组)
    - `/topic/:tid` 死代码 `'banned'` 字符串检查
    - `/unmoe` FE `result: string | number` → `string`
    - `/galgame/:gid/resource/all` BE `isLiked` hardcode false → 批量 `FindLikedSet`
    - `/galgame/:gid/comment/all` + `/comment/thread/:rootId` BE 缺 `IsLiked` + FE 类型 `number → string`
  - 第二组 (§6–§11)：
    - `/galgame-resource`, `/galgame-rating/all`, `/galgame-rating/:id` 路由组从 `api` 移到 `optAuth` (否则 `optionalUID` 恒 0，`isLiked` 永远 false)
    - `/galgame-official/:name` BE `OfficialDetail` 缺 `original`：wiki PR4/K-PR6 sub-change 加字段后 BE 漏接，FE 编辑模态打开时永远空白
    - `/galgame-tag/:name`, `/galgame-official/:name`, `/galgame-engine/:name` query rename helper 升级为多键 map (新增 `sortField`→`sort_field`、`sortOrder`→`sort_order`)，否则 wiki 静默 fallback 到默认排序
    - `/galgame-rating/:id` BE 缺 `galgameSeries`：FE Review JSON-LD 的 `isPartOf` 永远缺失；改为 wiki detail 拿 `series_id` 后再拉 `/series/:id` 的最小 brief
    - `/galgame-rating/all` FE 类型 `GalgameRatingCard` 缺 `short_summary` (BE 一直在返)
    - `/website`, `/website-category/:name`, `/website-tag/:name` BE 缺 SFW 过滤：FE Container.vue 宣称"默认仅显示 SFW 的网站"但 BE 全返 → 加 `age_limit='all'` scope，与 wiki content_limit 协议对齐
    - `/website/:domain/comment` 子节点孤儿处理：父评论作者被封后，原代码会把子评论提升到顶层（无 TargetUser 的悬挂回复），改为丢弃（与 galgame comment service 对齐）

---

## 0. 认证 / 身份

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /auth/me` | 🔒 | kungal-native (proxies OAuth `/auth/me`) | ⏳ | 返回当前登录态用户的 profile 投影；FE 一般用持久化 store 而非每次拉 |

## 1. 首页 / 动态 / 搜索 / Feed

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /home` | 🌐 | kungal-native + wiki batch | ⏳ | NSFW SFW filter via `IsSFW(c)` → `GetBatchPublic`；over-fetch 2× |
| `GET /activity` | 🌐 | kungal-native + wiki batch | ⏳ | 按 `type` 过滤的单类活动流；isSFW filters galgame name 注入 |
| `GET /activity/timeline` | 🌐 | kungal-native + wiki batch | ⏳ | 全类型时间线 |
| `GET /search` | 🌐 | kungal-native (Meilisearch via wiki for galgame) | ⏳ | type=galgame 时**强制 `content_limit=all`** (产品决策：搜索可发现 NSFW) |
| `GET /rss/topic` | 🌐 | kungal-native | ⏳ | SQL 强制 `is_nsfw=false` |
| `GET /rss/galgame` | 🌐 | kungal-native + wiki batch | ⏳ | 强制 `GetBatchPublic(..., true)` SFW |
| `GET /unmoe` | 🌐 | kungal-native | ⏳ | 站内违规记录列表 |
| `GET /admin/setting/register` | 🌐 | kungal-native | ⏳ | 公开读：当前注册策略 (是否开放、是否需邀请码) |

## 2. 话题 / 回复 / 投票

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /topic` | 🔐 | kungal-native | ⏳ | 话题列表；`isNSFW := !utils.IsSFW(c)` (list_repo SQL filter) |
| `GET /topic/:tid` | 🔐 | kungal-native | ⏳ | 话题详情；embed `bestAnswer` 给 FE JSON-LD；BE **不**做 SFW gate (由 FE 拦截) |
| `GET /topic/:tid/reply` | 🔐 | kungal-native | ⏳ | 回复列表；`GetReplies` 含 `isBestAnswer` flag |
| `GET /topic/:tid/reply/detail` | 🔐 | kungal-native | ⏳ | 单回复完整 (含 replyContentHtml 渲染) |
| `GET /topic/:tid/poll/topic` | 🔐 | kungal-native | ⏳ | 话题的投票题目 + 选项 |
| `GET /topic/:tid/poll/log` | 🔐 | kungal-native | ⏳ | 当前用户的投票记录 |
| `GET /resource` | 🌐 | kungal-native | ⏳ | 资源 section (g-seeking/g-other/t-help) 话题；同 `/topic` SFW 规则 |

## 3. Section / Category / 排行

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /section` | 🌐 | kungal-native | ⏳ | section 元数据 (g-galgame / t-tech 等) |
| `GET /category` | 🌐 | kungal-native | ⏳ | category → section 树 |
| `GET /ranking/galgame` | 🌐 | kungal-native + wiki batch | ⏳ | galgame 排行；`isSFW` → `GetBatchPublic` |
| `GET /ranking/topic` | 🌐 | kungal-native | ⏳ | topic 排行；SQL `t.is_nsfw=false` when SFW |
| `GET /ranking/user` | 🌐 | kungal-native | ⏳ | 用户排行；NSFW 无关 |

## 4. Galgame (核心列表 / 详情)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame` | 🌐 | kungal-native + wiki batch | ⏳ | 主列表；SFW filter via `GetBatchPublic(isSFW)`；含 type/language/platform/provider 本地 filter |
| `GET /galgame/:gid` | 🔐 | kungal-native + wiki direct | ⏳ | 详情；带 token 时 wiki 可返 owner 自己的 pending；BE **不** gate NSFW |
| `GET /galgame/mine` | 🔒 (`userAuth`) | proxy | ⏳ | `/galgame/mine` 透传 wiki；带 decline_reason |
| `GET /galgame/search/wizard` | 🔒 (`userAuth`) | kungal-native (proxy + token) | ⏳ | "发布向导"搜索；自动加 `?include_pending=true` |
| ~~`GET /galgame/check`~~ | — | — | ❌ removed | VNDB-ID 精确查重已被 name-based search 替代，FE 无调用方；K-PR 审计中已从 router.go 删除注册。 |
| `GET /galgame/:gid/resource/all` | 🔐 | kungal-native | ⏳ | 该 galgame 全部资源；FE 详情页用 |
| `GET /galgame/:gid/comment/all` | 🔐 | kungal-native | ⏳ | 该 galgame 评论 (含 root + 前 3 reply 预览) |
| `GET /galgame/:gid/comment/thread/:rootId` | 🔐 | kungal-native | ⏳ | 单条 root 评论的完整子树 |
| `GET /galgame/:gid/history/all` | 🔐 | wiki+local | ⏳ | revision 历史 (kungal-mapped camelCase) |
| `GET /galgame/:gid/pr/all` | 🔐 | wiki+local | ⏳ | PR 列表 (kungal-mapped) |
| `GET /galgame/:gid/link/all` | 🔐 | wiki+local | ⏳ | 链接列表 (kungal-mapped) |

## 5. Galgame Wiki 透传 (revisions / PR / links / aliases / contributors)

> 全部走 `GalgameWikiHandler.ProxyGet` → `path_mapper` → wiki service。响应 shape **不经 kungal 改写**，原样回传。

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame/:gid/revisions` | 🌐 | proxy | ⏳ | 历史版本列表 (wiki 直出) |
| `GET /galgame/:gid/revisions/:rev` | 🌐 | proxy | ⏳ | 单 revision 快照 |
| `GET /galgame/:gid/revisions/:rev/diff` | 🌐 | proxy | ⏳ | 单 revision diff |
| `GET /galgame/:gid/prs` | 🌐 | proxy | ⏳ | PR 列表 (与 `/pr/all` 区别：这条原样透传) |
| `GET /galgame/:gid/prs/:id` | 🌐 | proxy | ⏳ | 单 PR 详情 |
| `GET /galgame/:gid/links` | 🌐 | proxy | ⏳ | 链接 (原样透传) |
| `GET /galgame/:gid/aliases` | 🌐 | proxy | ⏳ | 别名列表 |
| ~~`GET /galgame/:gid/contributors`~~ | — | — | ❌ removed | FE 现从 detail 的 `contributor[]` 内嵌字段读取，独立列表无人调用；K-PR 审计中已从 router.go 删除注册。 |

## 6. Galgame 分类轴 (tag / official / engine / series)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame-tag` | 🌐 | kungal-native (forward to wiki) | ✅ | tag list；`isSFW` 时过滤 `category=sexual` |
| `GET /galgame-tag/search` | 🌐 | kungal-native (Meilisearch via wiki) | ✅ | tag 搜索；同上 sexual filter |
| `GET /galgame-tag/multi` | 🌐 | kungal-native + wiki batch | ✅ | 多 tag AND 搜 galgame |
| `GET /galgame-tag/:name` | 🌐 | kungal-native + wiki batch | 🔧 | tag detail；K-PR6 修复 `sortField/sortOrder` → snake_case 透传 (helper 多键 map 化) |
| `GET /galgame-official` | 🌐 | kungal-native (forward to wiki) | ✅ | official list；元数据无 NSFW filter |
| `GET /galgame-official/search` | 🌐 | kungal-native (Meilisearch via wiki) | ✅ | official 搜索 |
| `GET /galgame-official/:name` | 🌐 | kungal-native + wiki batch | 🔧 | official detail；K-PR6 补 `original` 字段 (wiki PR4 起新增，FE 编辑模态用) + sortField/Order 透传 |
| `GET /galgame-engine` | 🌐 | kungal-native (forward to wiki) | ✅ | engine list；元数据 |
| `GET /galgame-engine/:name` | 🌐 | kungal-native + wiki batch | 🔧 | engine detail；K-PR6 修复 sortField/sortOrder 透传 (与 tag/official 同一 helper) |
| `GET /galgame-series` | 🌐 | kungal-native (forward to wiki) | ✅ | series list；每条预载前 5 个 galgame sample |
| `GET /galgame-series/search` | 🌐 | proxy | ✅ | series 搜索 (Meilisearch via wiki，原样透传) |
| `GET /galgame-series/:id` | 🌐 | kungal-native + wiki batch | ✅ | series detail；**已删除 revision panel** (K-PR series-revision design)；含关联 galgame 列表 |

## 7. Galgame 分类轴 — Revisions (proxy 透传，K-PR5)

> 由 `router.go` 的 for-loop 动态注册：`for ent ∈ {galgame-tag, galgame-official, galgame-engine}`。
> `galgame-series` **已被故意排除** — series 成员变更落在 galgame 侧 revision，per-series 历史为空/误导，前端面板已下线，路由也未注册。

| 路径 (展开) | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame-tag/:id/revisions` | 🌐 | proxy | ✅ | tag 修订列表 |
| `GET /galgame-tag/:id/revisions/:rev` | 🌐 | proxy | ✅ | tag 单修订快照 |
| `GET /galgame-official/:id/revisions` | 🌐 | proxy | ✅ | official 修订列表 |
| `GET /galgame-official/:id/revisions/:rev` | 🌐 | proxy | ✅ | official 单修订快照 |
| `GET /galgame-engine/:id/revisions` | 🌐 | proxy | ✅ | engine 修订列表 |
| `GET /galgame-engine/:id/revisions/:rev` | 🌐 | proxy | ✅ | engine 单修订快照 |

## 8. Galgame 资源 (galgame-resource)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame-resource` | 🔐 | kungal-native + wiki batch | 🔧 | 资源列表；K-PR6 移到 `optAuth` 组 (此前 `api` 组导致 `isLiked` 永远 false)；`isSFW` → `GetBatchPublic`；`total` over-reports in SFW |
| `GET /galgame-resource/:id` | 🔐 | kungal-native + wiki | ✅ | 资源详情；BE **不** gate NSFW (由 FE 处理) |
| `GET /galgame-resource/:id/detail` | 🔐 | kungal-native + wiki | ✅ | 含完整下载链接 (会触发 download view 增量) |
| `GET /galgame-resource/:id/recommend` | 🔐 | kungal-native + wiki | ✅ | 相邻推荐资源 (同 galgame 的姊妹 resource) |

## 9. Galgame 评分 (galgame-rating)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame-rating/all` | 🔐 | kungal-native + wiki batch | 🔧 | 评分列表；K-PR6 移到 `optAuth` 组；FE 类型补 `short_summary`；BE 一直在返但 FE 类型缺 |
| `GET /galgame-rating/:id` | 🔐 | kungal-native + wiki | 🔧 | 单评分详情；K-PR6 移到 `optAuth` 组 + BE 补 `galgameSeries` brief (FE Review JSON-LD `isPartOf` 用) |

## 10. Galgame Submission Messages (wiki 消息流)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /galgame/messages/mine` | 🔒 (`authed`) | proxy (with Bearer) | ✅ | "我收到的 wiki 通知" (approved/declined/banned/unbanned 等) |
| `GET /galgame/messages/read-state` | 🔒 (`authed`) | kungal-native | ✅ | per-user HWM cursor (wiki_message_read_state) |
| `GET /admin/galgame/messages` | 🛡️ (`galgameAdmin`) | proxy (with Bearer) | ✅ | admin 审核队列 |

## 11. 网站 (website)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /website` | 🔐 | kungal-native | 🔧 | 网站列表；K-PR6 加 SFW filter (`age_limit='all'` scope) — 原 FE Container.vue 宣称"默认仅显示 SFW"但 BE 不过滤 |
| `GET /website/:domain` | 🔐 | kungal-native | ✅ | 网站详情；包含 comment 列表 |
| `GET /website/:domain/comment` | 🔐 | kungal-native | 🔧 | 该网站评论列表；K-PR6 修复孤儿子评论 (父被封时丢弃 reply 而非提升到顶层) |
| `GET /website-category/:name` | 🌐 | kungal-native | 🔧 | 网站分类下的所有网站；K-PR6 加 SFW filter |
| `GET /website-tag` | 🌐 | kungal-native | ✅ | 网站标签列表 |
| `GET /website-tag/:name` | 🌐 | kungal-native | 🔧 | 单标签详情；K-PR6 加 SFW filter |

## 12. 文档 (doc)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /doc/article` | 🌐 | kungal-native | ⏳ | 文档列表 (含分类/标签 facets) |
| `GET /doc/article/:slug` | 🌐 | kungal-native | ⏳ | 文档详情；含 `tagIds[]` (K-PR 修复，便于 rewrite 重填) |
| `GET /doc/category` | 🌐 | kungal-native | ⏳ | 文档分类列表 |
| `GET /doc/tag` | 🌐 | kungal-native | ⏳ | 文档标签列表 |

## 13. 工具集 (toolset)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /toolset` | 🔐 | kungal-native | ⏳ | 工具列表 |
| `GET /toolset/:id` | 🔐 | kungal-native | ⏳ | 工具详情 (含 practicality 评分聚合 + commentPreview) |
| `GET /toolset/:id/practicality` | 🔐 | kungal-native | ⏳ | 实用度评分分布 |
| `GET /toolset/:id/comment/all` | 🔐 | kungal-native | ⏳ | 工具评论列表 |
| `GET /toolset/:id/resource/detail` | 🌐 | kungal-native | ⏳ | 工具关联资源详情 |

## 14. 用户主页 (user/:id/*)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /user/status` | 🔒 (`userAuth`) | kungal-native | ⏳ | 当前用户的萌萌点/签到/未读数 (sum of message+system+chat) |
| `GET /user/:id` | 🌐 | kungal-native | ⏳ | 用户 profile (识别 `'banned'` 字面值) |
| `GET /user/:id/floating` | 🌐 | kungal-native | ⏳ | 浮动卡片 (hover user header 显示) |
| `GET /user/:id/galgames` | 🌐 | kungal-native + wiki batch | ⏳ | 用户上传/参与的 galgame；`isSFW` 过滤 |
| `GET /user/:id/galgame-comments` | 🌐 | kungal-native + wiki batch | ⏳ | 用户的 galgame 评论；按 galgame `contentLimit` 过滤 |
| `GET /user/:id/topics` | 🌐 | kungal-native | ⏳ | 用户话题；SQL `topic.is_nsfw=false` when SFW |
| `GET /user/:id/replies` | 🌐 | kungal-native | ⏳ | 用户回复；JOIN topic 过滤 NSFW topic 下的回复 |
| `GET /user/:id/comments` | 🌐 | kungal-native | ⏳ | 用户评论；JOIN topic 过滤 |
| `GET /user/:id/resources` | 🌐 | kungal-native + wiki batch | ⏳ | 用户上传的 galgame 资源 |
| `GET /user/:id/ratings` | 🌐 | kungal-native + wiki batch | ⏳ | 用户的 galgame 评分 |

## 15. 消息中心 / 聊天

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /message` | 🔒 (`authed`) | kungal-native | ⏳ | 个人通知列表 (user-to-user `message` 表) |
| `GET /message/admin` | 🔒 (`authed`) | kungal-native | ⏳ | 系统广播；含 per-user `isRead` (HWM cursor `system_message_read_state`) |
| `GET /message/nav/system` | 🔒 (`authed`) | kungal-native | ⏳ | nav-bar 概览 (notice + system unread count) |
| `GET /message/nav/contact` | 🔒 (`authed`) | kungal-native | ⏳ | 私信 nav (chat rooms 列表) |
| `GET /message/chat/history` | 🔒 (`authed`) | kungal-native | ⏳ | 单对话历史消息 |

## 16. 排行 / 更新日志

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /update/history` | 🌐 | kungal-native | ⏳ | GitHub commits feed 镜像 |
| `GET /update/todo` | 🌐 | kungal-native | ⏳ | 待办列表 (admin maintained) |

> 注：排行 (`/ranking/*`) 已在 §3 列出。

## 17. 管理后台 (admin)

| 路径 | 鉴权 | Handler 类型 | 状态 | 备注 |
|---|---|---|---|---|
| `GET /admin/overview/all` | ⚙️ (`admin`, role >= 3) | kungal-native | ⏳ | 后台数据总览 (user/topic/galgame/comment 增长) |
| `GET /admin/overview/stats` | ⚙️ (`admin`, role >= 3) | kungal-native | ⏳ | 后台统计数字快照 |
| `GET /admin/galgame/messages` | 🛡️ (`galgameAdmin`, role >= 2) | proxy | ⏳ | (重复列出，与 §10 一致) admin galgame 消息队列 |
| `GET /admin/setting/register` | 🌐 | kungal-native | ⏳ | (重复列出，与 §1 一致) 公开读注册策略 |

---

## 审计维度提示 (用于后续审核)

按 post.md / put.md 同款检查项逐条核对：

1. **路径与方法**：FE `useKunFetch` / `kunFetch` 中 URL / method 与 router.go 一致
2. **响应 shape**：FE 类型 `T` 与 BE DTO 字段名 / 嵌套结构一致 (camelCase vs snake_case)
3. **NSFW 过滤**：是否根据 `IsSFW(c)` 走 `GetBatchPublic`／`content_limit=sfw`
4. **认证语义**：optAuth / authed / admin 是否符合调用方期望 (匿名爬虫能看到吗？)
5. **错误码**：BE `ErrNotFound` (404 + code=233) 时 FE 是否优雅降级为 `KunNull` 而非崩
6. **分页 contract**：`{items, total}` vs `[]` 形态；`total` 是否被 wiki SFW filter 影响 (8 个已知端点)
7. **字段漂移**：post-K-PR refactor 后旧字段名残留 (`released`/`banner_image_hash`/`status` on system_message 等) — 见 [seo-audit.md §10.4](../seo-audit.md#104-字段名漂移每次后端重构都要核对此节)
8. **proxy vs kungal-native**：proxy 端点的字段 shape 取决于 wiki 当前版本；wiki 文档 (`docs/galgame_wiki/*.md`) 改了字段后 FE 是否同步

---

## 后续审计建议顺序

1. **公开高流量入口**：`/home`, `/galgame`, `/topic`, `/galgame-resource`, `/galgame-rating/all` — 影响 SEO + 热路径性能
2. **详情页 + JSON-LD 依赖**：`/galgame/:gid`, `/topic/:tid`, `/galgame-rating/:id`, `/toolset/:id`, `/website/:domain` — 字段缺失会直接打破 JSON-LD
3. **用户主页系列**：`/user/:id/*` — 字段名一致性 + NSFW JOIN 正确性
4. **proxy 透传组**：`/galgame/:gid/{revisions,prs,links,aliases,contributors}` 等 — wiki 字段 shape 变化的主要受害区
5. **admin / message** — 低流量但容易因 schema 演进 silent break

---

## 路径来源行号速查 (开发者参考)

`apps/api/internal/app/router.go`：

| 起始行 | 结束行 | 内容 |
|---|---|---|
| 22 | 28 | 首页 / 用户基础 |
| 36 | 54 | 用户主页系列 |
| 57 | 59 | 排行 |
| 62 | 63 | section / category |
| 66 | 69 | 文档 |
| 72 | 74 | 网站分类/标签 |
| 77 | 78 | 更新日志 |
| 81 | 81 | admin 公开读注册策略 |
| 84 | 85 | 活动 |
| 88 | 89 | 评分 |
| 92 | 92 | 资源 section topics |
| 95 | 95 | 搜索 |
| 98 | 99 | RSS |
| 102 | 102 | unmoe |
| 105 | 105 | 工具集 resource 详情 |
| 108 | 129 | galgame core (含 proxy revisions/prs/links/aliases/contributors) |
| 133 | 145 | galgame 分类轴 + galgame-resource list |
| 152 | 154 | galgame-resource 详情 (optAuth) |
| 157 | 162 | topic + reply + poll (optAuth) |
| 165 | 171 | galgame 子端点 (optAuth) |
| 174 | 176 | website (optAuth) |
| 179 | 182 | toolset (optAuth) |
| 189 | 189 | `/auth/me` (authed) |
| 221 | 228 | message + chat (authed) |
| 262 | 263 | wiki message mine + read-state (authed) |
| 349 | 351 | tag/official/engine revision proxy (for-loop) |
| 380 | 381 | admin overview (role>=3) |
| 393 | 393 | admin galgame messages (role>=2) |
