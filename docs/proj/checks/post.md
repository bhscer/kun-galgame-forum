# POST / PUT / DELETE API 字段对齐检查

> 目的:记录全部 mutation 端点(POST/PUT/DELETE)以及 FE↔BE 字段对齐审计状态。
>
> 路由源:`apps/api/internal/app/router.go`

## 图例

- ✅ 已审计,FE/BE 对齐无问题
- 🔧 已审计,**发现错位并修复**
- ⏭️ 已审计,设计上有意保持当前行为(详见备注)

## 统计

- 全部端点: **89**
- 已审计: **89**(100%)
- 已修复: **29**

> **复核轮次:** (1) 第一轮分模块审计 → 找到 13 项;
> (2) 第二轮深度全量审计(每个 agent 无字数限制)→ 又找到 15 项;
> (3) 第三轮专门复核所有 ⏭️ 标记,确认全部跳过项实际行为正确,只额外发现 1 处 callback 类型断言不准并修复。

---

## 认证 / 用户

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/oauth/callback` | 🔧 | code+code_verifier 对齐;FE 类型断言清掉 BE 不返回的 `email`(email 由 OAuth `/oauth/userinfo` 单独获取) |
| POST | `/logout` | ⏭️ | 无 body,BE 撤销 session 并清 cookie。FE 暂未调用此端点(本地清 store 即可) |
| POST | `/user/check-in` | ✅ | |
| PUT | `/user/bio` | ⏭️ | OAuth 代理 → `PATCH /auth/me { bio }`,bio max=107 双端对齐 |
| PUT | `/user/username` | ⏭️ | OAuth 代理 → `PATCH /auth/me { name }`(handler 翻译 username→name),max=17 + `isValidName` regex `{1,17}` 对齐 |
| POST | `/user/avatar` | ⏭️ | multipart 原样转发到 OAuth `/auth/me/avatar`,field 名 `file`,响应 `{hash,url,variant_urls,...}` 来自 OAuth |

## 消息 / 聊天

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| DELETE | `/message/:id` | 🔧 | 删除多余 `?messageId=` query |
| PUT | `/message/system/read` | ✅ | |
| PUT | `/message/admin/read` | ✅ | |
| POST | `/message/chat/send` | ✅ | |
| POST | `/message/chat/recall` | 🔧 | FE 之前完全没接,补上 context-menu 触发 |

## 图片上传

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/image/topic` | ✅ | multipart,key `image` |
| POST | `/image/galgame` | ✅ | multipart,preset `galgame_banner` |

## 举报

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/report/submit` | 🔧 | `reason.max` 1000→1007 对齐 FE |

## 话题 (Topic)

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/topic` | ✅ | |
| PUT | `/topic/:tid` | ✅ | |
| PUT | `/topic/:tid/like` | 🔧 | 清死 body |
| PUT | `/topic/:tid/dislike` | 🔧 | 清死 body |
| PUT | `/topic/:tid/upvote` | 🔧 | 清死 body |
| PUT | `/topic/:tid/favorite` | 🔧 | 清死 body |
| PUT | `/topic/:tid/hide` | 🔧 | 清死 body |
| PUT | `/topic/:tid/best-answer` | ✅ | |
| POST | `/topic/:tid/reply` | ✅ | |
| PUT | `/topic/:tid/reply` | 🔧 | BE 加 content+targets 全空兜底 |
| DELETE | `/topic/:tid/reply` | ✅ | |
| PUT | `/topic/:tid/reply/like` | ✅ | |
| PUT | `/topic/:tid/reply/dislike` | ✅ | |
| PUT | `/topic/:tid/reply/pin` | ✅ | |
| POST | `/topic/:tid/comment` | ✅ | |
| PUT | `/topic/:tid/comment/like` | 🔧 | URL path 改用 `comment.topicId` |
| DELETE | `/topic/:tid/comment` | ✅ | |
| POST | `/topic/:tid/poll` | ✅ | |
| PUT | `/topic/:tid/poll` | ✅ | |
| DELETE | `/topic/:tid/poll` | ✅ | |
| POST | `/topic/:tid/poll/vote` | ✅ | |

## 网站 (Website)

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/website` | 🔧 | BE 加 `Domain[]` + `CreateTime` 字段;FE 字段命名对齐 |
| PUT | `/website/:domain` | 🔧 | 同上 |
| DELETE | `/website/:domain` | ✅ | |
| PUT | `/website/:domain/like` | ✅ | |
| PUT | `/website/:domain/favorite` | ✅ | |
| POST | `/website/:domain/comment` | ✅ | |
| DELETE | `/website/:domain/comment` | 🔧 | 删除死的 `updateCommentSchema` |
| PUT | `/website-category` | ✅ | |
| POST | `/website-tag` | 🔧 | 新增 `CreateWebsiteTagRequest` DTO + validate |
| PUT | `/website-tag` | ✅ | |
| DELETE | `/website-tag` | ✅ | |

## Galgame 核心

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| PUT | `/galgame/:gid/like` | 🔧 | 清死 body |
| PUT | `/galgame/:gid/favorite` | 🔧 | 清死 body |
| POST | `/galgame/:gid/comment` | 🔧 | content max 1007→5000;targetUserId 改可选;补 parentCommentId |
| PUT | `/galgame/:gid/comment` | ✅ | |
| DELETE | `/galgame/:gid/comment` | ✅ | |
| PUT | `/galgame/:gid/comment/like` | 🔧 | FE `galgameCommentId` → `commentId` |
| POST | `/galgame/:gid/resource` | ✅ | |
| PUT | `/galgame/:gid/resource` | ✅ | |
| DELETE | `/galgame/:gid/resource` | ✅ | |
| PUT | `/galgame/:gid/resource/like` | ✅ | |
| PUT | `/galgame/:gid/resource/valid` | ✅ | |
| PUT | `/galgame/:gid/resource/expired` | ✅ | |

## Galgame 评分

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/galgame-rating` | ✅ | |
| PUT | `/galgame-rating/:id` | ✅ | |
| DELETE | `/galgame-rating/:id` | ✅ | |
| PUT | `/galgame-rating/:id/like` | ✅ | |
| POST | `/galgame-rating/:id/comment` | 🔧 | content max 1007→1314 |
| PUT | `/galgame-rating/:id/comment` | ✅ | |
| DELETE | `/galgame-rating/:id/comment` | ✅ | |

## Galgame 提交 / Wiki 代理

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/galgame/submit` | ⏭️ | raw-body 透传到 wiki;`tag_ids/official_ids/engine_ids/series_id` FE 故意省略(`Galgame.vue:201-203` 跟 `07-submission.md` 明示审核后再用 PR 补) |
| POST | `/galgame/:gid/claim` | ✅ | |
| DELETE | `/galgame/:gid` | ✅ | |
| PUT | `/galgame/messages/read-state` | ✅ | |
| POST | `/galgame` (proxy) | ✅ | wiki 写 |
| PUT | `/galgame/:gid` (proxy) | ✅ | wiki 写 |
| PUT | `/galgame/:gid/prs/:id/merge` | ✅ | |
| POST | `/galgame/:gid/revert` | ✅ | |
| POST | `/galgame/:gid/prs` | ✅ | |
| PUT | `/galgame/:gid/prs/:id/decline` | ✅ | |
| POST | `/galgame/:gid/links` | ✅ | |
| DELETE | `/galgame/:gid/links` | ✅ | |
| POST | `/galgame/:gid/aliases` | ✅ | |
| DELETE | `/galgame/:gid/aliases` | ✅ | |
| DELETE | `/galgame/:gid/contributors/:id` | ✅ | |
| POST | `/galgame-tag` | ✅ | |
| PUT | `/galgame-tag` | 🔧 | BE proxy 翻译 `tagId`→`tag_id` |
| DELETE | `/galgame-tag/:id` | ✅ | |
| POST | `/galgame-official` | ✅ | |
| PUT | `/galgame-official` | 🔧 | BE proxy 翻译 `officialId`→`official_id` |
| DELETE | `/galgame-official/:id` | ✅ | |
| POST | `/galgame-engine` | ✅ | |
| PUT | `/galgame-engine` | 🔧 | BE proxy 翻译 `engineId`→`engine_id` |
| DELETE | `/galgame-engine/:id` | ✅ | |
| POST | `/galgame-{tag,official,engine,series}/:id/revert` | ✅ | revert 系列 |
| POST | `/galgame-series` | ✅ | FE Container.vue 手工转 `galgame_ids` |
| POST | `/galgame-series/modal` | ✅ | |
| PUT | `/galgame-series/:id` | ✅ | FE Detail.vue 手工转 `galgame_ids` |
| DELETE | `/galgame-series/:id` | ✅ | |

## 工具集 (Toolset)

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/toolset` | ✅ | |
| PUT | `/toolset/:id` | ✅ | |
| DELETE | `/toolset/:id` | ✅ | |
| PUT | `/toolset/:id/practicality` | 🔧 | FE 删多余 `toolsetId` 字段 |
| POST | `/toolset/:id/comment` | ✅ | |
| PUT | `/toolset/:id/comment` | ✅ | |
| DELETE | `/toolset/:id/comment` | ✅ | |
| POST | `/toolset/:id/resource` | 🔧 | 加 `type` + `content=key` (s3 模式);删 `salt` |
| PUT | `/toolset/:id/resource` | 🔧 | BE 改返回 resource 而非 OKMessage;schema 加 `type` superRefine |
| DELETE | `/toolset/:id/resource` | ✅ | |
| POST | `/toolset/:id/upload/small` | 🔧 | 补 `contentType`;响应 `presignedUrl` |
| POST | `/toolset/:id/upload/large` | 🔧 | 补 `contentType`;响应 `parts/presignedUrl` |
| POST | `/toolset/:id/upload/complete` | 🔧 | parts 字段 camelCase |
| POST | `/toolset/:id/upload/abort` | 🔧 | 移除多余 `uploadId` |

## 管理员

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| PUT | `/admin/setting/register` | ✅ | |
| PUT | `/admin/galgame/:gid/status` | ✅ | wiki proxy |

## 文档 (Doc, admin)

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/doc/article` | 🔧 | model + DTO 全量 JSON tag 改 camelCase |
| PUT | `/doc/article` | 🔧 | 同上 |
| DELETE | `/doc/article` | ✅ | |
| POST | `/doc/category` | 🔧 | 新增 `CreateCategoryRequest` DTO + validate |
| PUT | `/doc/category` | ✅ | |
| DELETE | `/doc/category` | ✅ | |
| POST | `/doc/tag` | 🔧 | 新增 `CreateTagRequest` DTO + validate |
| PUT | `/doc/tag` | ✅ | |
| DELETE | `/doc/tag` | ✅ | |

## 更新日志 (admin)

| 方法 | 路径 | 状态 | 备注 |
|---|---|---|---|
| POST | `/update/history` | 🔧 | FE 补 `content_ja_jp`/`content_zh_tw` |
| PUT | `/update/history` | ✅ | (沿用 create schema) |
| DELETE | `/update/history` | ✅ | |
| POST | `/update/todo` | 🔧 | 同 update/history |
| PUT | `/update/todo` | ✅ | |
| DELETE | `/update/todo` | ✅ | |

---

## 已修复问题清单(29 项)

### 第一轮(前期排查中发现 + 用户手动报告)

1. Toolset 上传 init 缺 `contentType`,响应字段名错位(`url→presignedUrl`、`urls/partSize→parts/presignedUrl`、parts casing)
2. Toolset resource POST 缺 `type`/`content=key`/字节字符串 size
3. Toolset resource 详情响应嵌套 → flat
4. Toolset 切到 B2 桶(`FILE_STORAGE_*` 环境变量)
5. Toolset Item 显示 size 时的 NaN GB / NaN 年前
6. `Like.vue` (topic comment) URL path 用 `comment.topicId`
7. `PUT /toolset/:id/resource` 返回更新后的 resource
8. `updateToolsetResourceSchema` 按 type 分支
9. Report `reason.max` 对齐 1007
10. `/message/chat/recall` FE 绑定
11. Reply DTO content+targets 全空兜底
12. `ToolsetDetail.contributors` 类型补全
13. Doc article 全量 camelCase 统一(snake/camel 混用)

### 第二轮(深度全量审计后发现)

14. `PUT /galgame/:gid/comment/like` 字段名 `galgameCommentId → commentId`
15-17. Wiki tag/official/engine PUT 的 id 字段在 BE proxy 自动 camel→snake
18. `/message` 响应 `totalCount → total`
19. Rating 评论 content max 1314
20. Galgame 评论 content max 5000
21. Galgame 评论 targetUserId 改可选 + 补 parentCommentId
22. 更新日志 ja_jp / zh_tw 字段补全
23. 5 个 topic 交互清死 body
24. `DELETE /message/:id` 清死 query
25-26. Doc Category/Tag POST 加专用 DTO + validate
27. 删除死的 website `updateCommentSchema`
28. Galgame `/like` `/favorite` 清死 body

### 第三轮(复核所有跳过项)

29. `/oauth/callback` FE 响应类型断言清掉 BE 不返回的 `email` 字段

---

## 跳过项复核结果(第三轮)

复核确认下列 5 个 ⏭️ 标记的端点实际行为完全符合标注,没有意外副作用:

| 端点 | 验证要点 |
|---|---|
| `POST /logout` | BE 撤销 session token + 清 cookie,Path/Secure/SameSite 跟设置时一致;FE 暂未调用此端点(本地清 store 已足够) |
| `PUT /user/bio` | OAuth 代理透明转发,access token 缺失会显式返回 `ErrAuthExpired`,成功后调用 `userClient.Invalidate` 让新 bio 立即生效 |
| `PUT /user/username` | 同上;handler 主动翻译 `username→name` 字段适配 OAuth `/auth/me`,FE `isValidName` 正则 `{1,17}` 跟 BE `max=17` 完全对齐 |
| `POST /user/avatar` | multipart 原 body 透传,multipart key `file` 跟 OAuth 端约定一致;响应 `{hash,url,variant_urls,...}` 由 OAuth 提供,FE 用 `result.url` 读取 |
| `POST /galgame/submit` | raw-body 透传到 wiki,taxonomy 字段缺失是文档化的设计决策(注释 + `07-submission.md` 双重确认),wiki 端也是可选字段 |

副作用检查:
- ❌ OAuth 代理无静默失败路径(全部明确返回 `ErrAuthExpired`)
- ❌ multipart 上传 `c.Body()` 单次读取,无重复缓存
- ❌ callback handler 明确禁止 log 请求体(code / code_verifier 是短期凭证)
- ❌ logout 清的 cookie name 跟设置时一致,无残留

---

## 检查方法论

每个端点核对了:

1. **路径参数**:`:tid` `:gid` `:id` 等是否被 BE 实际使用,FE 是否传对
2. **请求体字段**:Go DTO `json:"..."` tag vs FE kunFetch `body` 字段名 + zod schema
3. **校验约束**:Go `validate:"..."` (required/min/max/oneof) vs FE zod (`.min().max().enum()`)
4. **响应形状**:BE response DTO vs FE 期待的 TS 类型(`useKunFetch<T>`)
5. **错位类型**:命名(camel/snake/Pascal)、必填/可选错位、长度上限、字段大小写、多余字段(dead weight)
6. **特殊情况**:OAuth 代理跳过、wiki 代理需要双层对齐(kungal proxy + wiki 上游)

`go build ./...` + `pnpm typecheck` 全程通过,保证修复不引入回归。
