<script setup lang="ts">
defineProps<{
  galgame: GalgameDetail
}>()

const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const isOpening = ref(false)

const handleRewriteGalgame = async (galgame: GalgameDetail) => {
  if (isOpening.value) return
  isOpening.value = true

  // Wiki PR semantics are REPLACE-ALL for aliases / tag_ids /
  // official_ids / engine_ids / links: whatever we submit becomes the
  // entire set. So the working copy must be hydrated COMPLETELY from
  // current data, or a save silently wipes anything we left out.
  //
  // tag/official/engine come embedded in the detail payload already;
  // links do NOT — fetch them explicitly. If that fetch fails we keep
  // links empty but loudly warn, so the user can re-confirm the links
  // section instead of unknowingly clearing every link on submit.
  const linkRes = await kunFetch<GalgameLink[]>(
    `/galgame/${galgame.id}/link/all`,
    { method: 'GET' }
  )
  if (linkRes === null) {
    useMessage(
      '未能加载该 Galgame 当前的相关链接, 提交前请在「相关链接」区核对, 否则可能清空已有链接',
      'warn',
      5000
    )
  }

  galgamePR.value[0] = {
    id: galgame.id,
    vndbId: galgame.vndbId,
    name: galgame.name,
    // Editor (type="rewrite") edits raw markdown — use `markdown`, NOT
    // `introduction` (that field is pre-rendered HTML).
    introduction: galgame.markdown,
    contentLimit: galgame.contentLimit as 'sfw' | 'nsfw',
    ageLimit: galgame.ageLimit,
    originalLanguage: galgame.originalLanguage as Language,
    alias: [...galgame.alias],
    tags: [...galgame.tag],
    officials: [...galgame.official],
    engines: [...galgame.engine],
    links: (linkRes ?? []).map((l) => ({ name: l.name, link: l.link })),
    note: ''
  }

  isOpening.value = false
  // ?type=pr arms the `prevent` middleware: on a hard refresh / direct
  // hit the temp store is empty (persist:false) → it redirects out
  // instead of letting the form crash on galgamePR[0]!.
  await navigateTo('/edit/galgame/rewrite?type=pr')
}
</script>

<template>
  <KunTooltip text="编辑">
    <KunButton
      :is-icon-only="true"
      variant="light"
      color="default"
      size="lg"
      :loading="isOpening"
      :disabled="isOpening"
      @click="handleRewriteGalgame(galgame)"
    >
      <KunIcon name="lucide:pencil" />
    </KunButton>
  </KunTooltip>
</template>
