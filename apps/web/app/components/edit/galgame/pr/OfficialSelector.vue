<script setup lang="ts">
// Developer/publisher (会社) picker. /galgame-official/search expects a
// space-tokenized `q` array (same convention as
// components/galgame/official/Container.vue). Binds full
// GalgameOfficialItem objects; Footer derives official_ids[].
const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const model = computed<GalgameOfficialItem[]>({
  get: () => galgamePR.value[0]?.officials ?? [],
  set: (v) => {
    if (galgamePR.value[0]) galgamePR.value[0].officials = v
  }
})

const search = async (q: string): Promise<GalgameOfficialItem[]> => {
  const res = await kunFetch<GalgameOfficialItem[]>(
    '/galgame-official/search',
    { method: 'GET', query: { q: q.split(' ').filter(Boolean) } }
  )
  return res ?? []
}
</script>

<template>
  <EditGalgamePrEntitySelectorBase
    v-model="model"
    title="会社"
    description="开发商 / 发行商。搜索 Wiki 已有会社并添加, 提交后整组替换。"
    placeholder="搜索会社 (空格分隔多个关键字)"
    :search="search"
  />
</template>
