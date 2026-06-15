# 老图迁移到 image_service：kungal 侧清单与迁移（2026-06-15）

> 本仓自有工程笔记（**非** infra 镜像）。背景：OAuth 侧正在跑**用户头像**迁移（老头像 → image_service）。
> 本文盘清 **kungal 自己的数据里还有哪些老图**需要迁移，并给出统计 SQL 与迁移命令。

## 0. 结论：头像在 kungal 这边零工作

`users.avatar` 早在 **migration 007** 就删了——kungal **不持久化头像**，每次从 OAuth 实时取。
OAuth 把头像迁完，kungal 自动拿到新 URL。**头像不需要、也无法在 kungal 侧迁。**

kungal 真正的「老图」是另一类：**用户内容里嵌的老图床图片**。

## 1. 老 vs 新 的判别

| | 形态 | 例子 |
|---|---|---|
| **老（需迁移）** | 路径式、非内容寻址，老图床主机 | `https://image.kungal.com/topic/user_38351/茅羽耶-1761790962059.webp` |
| **新（image_service）** | 内容寻址，两级分片 | `https://image.kungal.iloveren.link/{aa}/{bb}/{hash}.webp` |

判别子：内容里出现 **`image.kungal.com`** = 老图（老图床早已停写，新上传一律走 `image.kungal.iloveren.link`）。

## 2. 清单（数据来自 postgres-kungalgame MCP **缓存快照，非 live prod**——精确值请用 §3 在线上重跑）

**需要迁移的：**

| 来源（content 列） | 含老图的行数 | 老图引用数 |
|---|---|---|
| `topic.content` | ~1298 | ~4777 |
| `topic_reply.content` | ~488 | ~659 |
| `chat_message.content`（裸 URL，部分已 404） | ~29 | ~29 |
| `galgame_comment.content` | ~1 | ~1 |
| **合计** | **~1816 行** | **~5466 处**（去重后文件数更少，同图常被多次引用） |

**确认不用迁的：**
- **头像**：见 §0。`chat_room.avatar` 1148 行全空；`doc_category.icon` 全空。
- `doc_article.banner`（29 行）：都是仓库静态资源 `/content/.../banner.avif`，非上传图。
- `galgame_website.icon`：基本是**外站 favicon**（bgm.tv / ggbases…），非我方图。
- galgame 封面/截图：早已 image_service（`image_hash`）。
- `sticker.kungal.com`（topic 156 + reply 409）：表情贴独立服务，语义上不算老图床上传——**是否纳入由产品定**。
- 第三方外链（raw.githubusercontent / postimg / ymgal…）：非我方图，迁不了（热链，可能失效，另一回事）。

> ⚠️ 部分 `image.kungal.com` 图**已经 404**（chat 样本里直接带「404错误」文字）——老图床在退场，**迁移要趁早**，且要容忍取不到原图。

## 3. 在 live prod 重跑的统计 SQL

```sql
-- 每表含老图的行数 + 已在新图床的行数
SELECT 'topic'           AS tbl, count(*) FILTER (WHERE content LIKE '%image.kungal.com%')          AS rows_old,
                                  count(*) FILTER (WHERE content LIKE '%image.kungal.iloveren.link%') AS rows_new FROM topic
UNION ALL SELECT 'topic_reply',     count(*) FILTER (WHERE content LIKE '%image.kungal.com%'),          count(*) FILTER (WHERE content LIKE '%image.kungal.iloveren.link%') FROM topic_reply
UNION ALL SELECT 'chat_message',     count(*) FILTER (WHERE content LIKE '%image.kungal.com%'),          count(*) FILTER (WHERE content LIKE '%image.kungal.iloveren.link%') FROM chat_message
UNION ALL SELECT 'galgame_comment',  count(*) FILTER (WHERE content LIKE '%image.kungal.com%'),          count(*) FILTER (WHERE content LIKE '%image.kungal.iloveren.link%') FROM galgame_comment;

-- 去重老图文件数（迁移工作量 = 这个数；同图多引用只传一次）
SELECT count(DISTINCT url) AS distinct_old_files FROM (
  SELECT (regexp_matches(content, 'https://image\.kungal\.com/[^\s)"''>\]]+', 'g'))[1] AS url FROM topic
  UNION ALL SELECT (regexp_matches(content, 'https://image\.kungal\.com/[^\s)"''>\]]+', 'g'))[1] FROM topic_reply
  UNION ALL SELECT (regexp_matches(content, 'https://image\.kungal\.com/[^\s)"''>\]]+', 'g'))[1] FROM chat_message
  UNION ALL SELECT (regexp_matches(content, 'https://image\.kungal\.com/[^\s)"''>\]]+', 'g'))[1] FROM galgame_comment
) s;
```

## 4. 迁移命令：`cmd/backfill-content-images`

按现有 `backfill-friend-link-banners` 同一套路实现（config → image client → DB → dry-run → 抓原图 → 传 image_service → 改写 DB）。

逐个去重老 URL：HTTP 抓 `image.kungal.com` 原图 → 用 `topic` preset 传 image_service → 缓存新 URL → 把每行 content 里的旧 URL 全替换成新 URL。**跨表共享缓存**（同图只传一次）。抓不到 / 404 的 URL **记日志并跳过**（该行保留旧 URL，重跑会再试）。**只写 `content` 列，绝不动 `updated`**（否则会把老帖顶到「最近更新」）。

```bash
docker compose -f docker-compose.prod.yml --profile jobs run --rm tools \
  backfill-content-images                     # dry-run（默认）：只报告去重老图数 / 可抓取数 / 404 数
  backfill-content-images -dry-run=false      # 真跑：重托管 + 改写
  backfill-content-images -dry-run=false -limit=20   # 先拿 20 行/表 试水
  backfill-content-images -base=http://<内网镜像>     # 老图床若从 job 容器访问不到，改抓取来源
```

**安全性**：`-dry-run` 默认 **TRUE**，不发请求、不写库。幂等——改写后 content 不再匹配 `%image.kungal.com%`，重跑自动跳过。

**前提与注意**：
1. **数字先在线上重跑确认**（本仓快照非 live prod）。
2. **老图床可达性**：`backfill-friend-link-banners` 注释提过「公网域名在集群内不 hairpin、内网 web 端口不通」。若 job 容器抓不到 `image.kungal.com`，用 `-base` 指到内网镜像 / S3 端点；若老图存在对象存储而非 HTTP，需改 `fetch`（本命令假设 HTTP 可达）。
3. **趁早**：已有 404，老图床在退场。
4. 跑前 **dry-run + 小 `-limit` 试水**，再全量。
