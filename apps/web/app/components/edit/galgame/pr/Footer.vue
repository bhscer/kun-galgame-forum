<script setup lang="ts">
import { updateGalgameSchema } from '~/validations/galgame'

const { galgamePR } = storeToRefs(useTempGalgamePRStore())

// Creator/admin edit directly (instant); others open a PR. Drives the
// button label here and the endpoint branch in the handler.
const canDirectEdit = computed(
  () => galgamePR.value[0]?.canDirectEdit ?? false
)

const isPublishing = ref(false)

const handlePublishGalgamePR = async () => {
  const galgame = galgamePR.value[0]
  if (!galgame) return

  // Wire-format payload uses snake_case keys to match the wiki PR API
  // (POST /galgame/:gid/prs). Aliases are sent as a string array
  // (replace-all semantics), unlike the create endpoint which takes a
  // comma-separated string — see docs/galgame_wiki/02-revisions-and-prs.md.
  const aliasList = Array.isArray(galgame.alias)
    ? galgame.alias.filter((a) => a.trim().length > 0)
    : String(galgame.alias)
        .split(',')
        .map((a) => a.trim())
        .filter((a) => a.length > 0)

  // Relation arrays are sent as id arrays / {name,link} objects and
  // REPLACE the whole set on merge — Rewrite.vue pre-hydrated them from
  // current data so an untouched section round-trips unchanged. Blank
  // link rows are dropped here so the schema's min(1) doesn't reject
  // them.
  const links = galgame.links
    .map((l) => ({ name: l.name.trim(), link: l.link.trim() }))
    .filter((l) => l.name.length > 0 && l.link.length > 0)

  const data: Record<
    string,
    number | string | string[] | number[] | { name: string; link: string }[]
  > = {
    vndb_id: galgame.vndbId,
    name_en_us: galgame.name['en-us'],
    name_ja_jp: galgame.name['ja-jp'],
    name_zh_cn: galgame.name['zh-cn'],
    name_zh_tw: galgame.name['zh-tw'],
    intro_en_us: galgame.introduction['en-us'],
    intro_ja_jp: galgame.introduction['ja-jp'],
    intro_zh_cn: galgame.introduction['zh-cn'],
    intro_zh_tw: galgame.introduction['zh-tw'],
    content_limit: galgame.contentLimit,
    age_limit: galgame.ageLimit,
    original_language: galgame.originalLanguage,
    aliases: aliasList,
    tag_ids: galgame.tags.map((t) => t.id),
    official_ids: galgame.officials.map((o) => o.id),
    engine_ids: galgame.engines.map((e) => e.id),
    links,
    note: galgame.note
  }

  const result = updateGalgameSchema.safeParse(data)
  if (!result.success) {
    const message = JSON.parse(result.error.message)[0]
    useMessage(
      `位置: ${message.path[0]} - 错误提示: ${message.message}`,
      'warn'
    )
    return
  }
  // Creator / admin edit directly (PUT /galgame/:gid — instant, new
  // revision). Everyone else opens a PR (POST /:gid/prs). Decided in
  // Rewrite.vue at hydration; see GalgameEditStoreTemp.canDirectEdit.
  const direct = canDirectEdit.value

  const res = await useComponentMessageStore().alert(
    direct ? '确定直接保存对该 Galgame 的修改吗?' : '确定发布 Galgame 信息更新请求吗?',
    direct
      ? '保存后立即生效并写入一条新的版本历史。别名 / 标签 / 会社 / 引擎 / 相关链接为整组替换, 请确认上方各项即为最终完整列表。'
      : '别名 / 标签 / 会社 / 引擎 / 相关链接为整组替换, 请确认上方各项即为最终完整列表。'
  )
  if (!res) {
    return
  }

  if (isPublishing.value) {
    return
  } else {
    isPublishing.value = true
  }

  // 可选 banner：用户在 PR 页改了 banner 就一并提交，没改则只发 JSON 字段。
  // multipart 约定见 docs/galgame_wiki/api-reference.md "Banner 上传"。
  const banner = await getImage('kun-galgame-publish-banner')

  // Direct edit → PUT /galgame/:gid; PR → POST /galgame/:gid/prs. Both
  // accept the same JSON body or multipart (data + file) per wiki
  // 01-galgame.md §multipart.
  const path = direct
    ? `/galgame/${galgame.id}`
    : `/galgame/${galgame.id}/prs`
  const method = direct ? 'PUT' : 'POST'

  let response: unknown
  if (banner instanceof Blob) {
    const formData = new FormData()
    formData.append('data', JSON.stringify(data))
    formData.append('file', banner)
    response = await kunFetch(path, { method, body: formData })
  } else {
    response = await kunFetch(path, { method, body: data })
  }
  isPublishing.value = false

  if (response) {
    if (banner instanceof Blob) {
      await deleteImage('kun-galgame-publish-banner')
    }
    // PUT /galgame/:gid does NOT process aliases/links (01-galgame.md);
    // they live on dedicated add/remove endpoints. The PR endpoint DOES
    // take them (replace-all), so only the direct path needs to
    // reconcile aliases/links itself against the current set.
    if (direct) {
      await reconcileAliasesLinks(galgame.id, galgame.alias, galgame.links)
    }
    useKunLoliInfo(direct ? '已保存, 修改已生效' : '创建更新请求成功', 5)
    await navigateTo(`/galgame/${galgame.id}`, {
      replace: true
    })
  }
}

// Reconcile aliases & links to the edited set via the dedicated
// per-item endpoints (docs 03-relations): GET current → POST the added,
// DELETE the removed (each creates its own revision). Best-effort: a
// failure warns but doesn't undo the already-applied PUT.
const reconcileAliasesLinks = async (
  gid: number,
  desiredAliasRaw: string[],
  desiredLinksRaw: { name: string; link: string }[]
) => {
  let failed = false

  // aliases
  const curAliases =
    (await kunFetch<{ id: number; name: string }[]>(
      `/galgame/${gid}/aliases`,
      { method: 'GET' }
    )) ?? []
  const desiredAlias = [
    ...new Set(desiredAliasRaw.map((a) => a.trim()).filter(Boolean))
  ]
  for (const name of desiredAlias) {
    if (!curAliases.some((a) => a.name === name)) {
      const r = await kunFetch(`/galgame/${gid}/aliases`, {
        method: 'POST',
        body: { name }
      })
      if (r === null) failed = true
    }
  }
  for (const a of curAliases) {
    if (!desiredAlias.includes(a.name)) {
      const r = await kunFetch(`/galgame/${gid}/aliases`, {
        method: 'DELETE',
        body: { id: a.id }
      })
      if (r === null) failed = true
    }
  }

  // links — identity is the (name, link) pair
  const curLinks =
    (await kunFetch<{ id: number; name: string; link: string }[]>(
      `/galgame/${gid}/links`,
      { method: 'GET' }
    )) ?? []
  const key = (n: string, l: string) => JSON.stringify([n, l])
  const desiredLinks = desiredLinksRaw
    .map((l) => ({ name: l.name.trim(), link: l.link.trim() }))
    .filter((l) => l.name && l.link)
  const desiredSet = new Set(desiredLinks.map((l) => key(l.name, l.link)))
  const curSet = new Set(curLinks.map((l) => key(l.name, l.link)))
  for (const l of desiredLinks) {
    if (!curSet.has(key(l.name, l.link))) {
      const r = await kunFetch(`/galgame/${gid}/links`, {
        method: 'POST',
        body: { name: l.name, link: l.link }
      })
      if (r === null) failed = true
    }
  }
  for (const l of curLinks) {
    if (!desiredSet.has(key(l.name, l.link))) {
      const r = await kunFetch(`/galgame/${gid}/links`, {
        method: 'DELETE',
        body: { id: l.id }
      })
      if (r === null) failed = true
    }
  }

  if (failed) {
    useMessage('部分别名 / 链接未能保存, 请到该条目核对后重试', 'warn', 5000)
  }
}
</script>

<template>
  <div class="flex justify-end">
    <KunButton
      :disabled="isPublishing"
      :loading="isPublishing"
      size="lg"
      @click="handlePublishGalgamePR"
    >
      {{ canDirectEdit ? '保存修改' : '提交更新请求' }}
    </KunButton>
  </div>
</template>
