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

## 2. 清单（**live prod kungalgame 库实测**，infra 2026-06-15 确认；与缓存快照一致——老图床早停写、数量稳定）

**需要迁移的：**

| 来源（content 列） | 含老图行数 | 已在 image_service |
|---|---|---|
| `topic.content` | 1,298 | 26 |
| `topic_reply.content` | 488 | 32 |
| `chat_message.content`（裸 URL，部分已 404） | 29 | 12 |
| `galgame_comment.content` | 1 | 8 |
| **合计** | **1,816 行** | — |

**实际工作量 = 5,440 个去重老图文件**（同图常被多处引用，所以文件数 < 引用数）。
可行性 infra 已确认：老图床走 Cloudflare（非本集群入口，无 hairpin 问题），dokploy-network 内随机抽 60 张 100% 返回 200；kungal client 配额 10,000/天·10GB/天，~5,440 张(~270MB)一轮跑完绰绰有余。

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

逐个去重老 URL：HTTP 抓 `image.kungal.com` 原图 → 用 `topic` preset 传 image_service → 缓存新 URL → 把每行 content 里的旧 URL 全替换成新 URL。**跨表共享缓存**（同图只传一次）。抓不到 / 404 的 URL **记日志并跳过**（该行保留旧 URL，重跑会再试）。**只写 `content` 列，绝不动 `updated`**（否则会把老帖顶到「最近更新」）。顺序执行，~5,440 张约 **1 小时**。

**审计轨迹（就是 job 日志）**：每张 `已重托管 old→new`、每个取不到的 `失败/404` URL、每行 `已改写 table id 替换处数`，跑完一条总结 + 失败 URL 列表。导出运行日志即留底（可据此核对 / 回滚 / 跟进死图）。

```bash
docker compose -f docker-compose.prod.yml --profile jobs run --rm tools \
  backfill-content-images                     # dry-run（默认）：纯统计去重老图文件数 + 各表行数（不联网、不写库）
  backfill-content-images -dry-run=false -limit=20   # 先拿 20 行/表 试水（真抓真传真改）
  backfill-content-images -dry-run=false      # 全量：重托管 + 改写（~1h）
  backfill-content-images -base=http://<内网镜像>     # 老图床若从 job 容器访问不到，改抓取来源（infra 实测可达，一般不用）
```

**安全性**：`-dry-run` 默认 **TRUE**——只 scan + 报工作量，**不联网、不写库**。幂等——改写后 content 不再匹配 `%image.kungal.com%`，重跑自动跳过（只重试上次的死图）。只改 `content`，不顶帖。

**分级流程**（同头像迁移）：
1. 部署带本命令的 forum **tools 镜像**（`fb052342` 已提交，但还没进已部署镜像）。
2. **dry-run** 确认去重文件数 ≈ 5,440。
3. **`-limit=20` 小批试水**（真改 20 行/表），抽查改写结果与新图能打开。
4. **全量** `-dry-run=false`，导出日志留底。
5. 跑完核对：死图列表人工跟进；`SELECT count(*) … LIKE '%image.kungal.com%'` 应只剩死图行。

**注意**：① 老图床可达性 infra 已实测 OK（Cloudflare，非集群入口，无 hairpin）；若环境变化抓不到，用 `-base` 指内网镜像。② **趁早**——已有 404，老图床在退场。③ 表情贴 `sticker.kungal.com` **不在范围**（独立服务，是否纳入另议）。

## 5. 执行记录（2026-06-15 完成）

按上面分级流程在 prod 跑完（`docker compose --profile jobs run --rm tools backfill-content-images`）：

- **dry-run**：去重老图 5,440 张，与 infra live prod 实测一致。
- **`-limit=20` 试水**：重托管 96 张 / 改写 61 行 / 0 失败；抽查新 URL 200、剩余数恰好减 20。
- **全量**（13:18→15:18 UTC，约 2 小时，顺序）：**重托管 5,342 张、改写 1,747 行（5,370 处）、0 报错**。
- **死图 3 张**（404，已保留旧 URL，本就已坏，非回归）：
  - `image.kungal.com/kun`（残缺 URL，非真实图）
  - `image.kungal.com/topic/user_10050/Clobber1238-1722604366830.webp`（老图床上已删）
  - `image.kungal.com/topic/user_2/鲲-1780457802051.webp1`（文件名带错字 `1`）
- **收尾核对**：剩余含 `image.kungal.com` 的行 = topic 1 / reply 1 / chat 6 = **8 行**，与「跳过 8 行」吻合，且全部只引用上述 3 张死图（`NOT /kun AND NOT Clobber1238` 过滤后为 0）。新图 200 可访问。
- 表情贴 `sticker.kungal.com` 未动（按约定）。审计全量日志见 prod `/tmp/bcimg-full.log`。
- **死图清理（2026-06-15）**：3 张死图中,reply 3354 的 markdown 图 + 6 条 chat 的裸 URL 已删除（共 7 行,只删图不删消息/行）；topic 1701 的 `image.kungal.com` 是帖子里的**配置示例文本**（`IMAGE_BED_ENDPOINT = "…"`），非图片,**保留**。至此 content 内已无任何坏图引用。表情贴 `sticker.kungal.com` 仍在用,不动。
