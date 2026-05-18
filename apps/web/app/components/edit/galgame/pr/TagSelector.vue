<script setup lang="ts">
// Tag picker for the PR edit form. Network search via
// /galgame-tag/search (kungal proxies wiki). Binds the full
// GalgameTagItem objects into the temp PR store; Footer derives
// tag_ids[] at submit.
const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const model = computed<GalgameTagItem[]>({
  get: () => galgamePR.value[0]?.tags ?? [],
  set: (v) => {
    if (galgamePR.value[0]) galgamePR.value[0].tags = v
  }
})

const search = async (q: string): Promise<GalgameTagItem[]> => {
  const res = await kunFetch<GalgameTagItem[]>('/galgame-tag/search', {
    method: 'GET',
    query: { q }
  })
  return res ?? []
}
</script>

<template>
  <EditGalgamePrEntitySelectorBase
    v-model="model"
    title="标签"
    description="输入关键字搜索 Wiki 已有标签, 点击添加。提交后将整组替换该 Galgame 的标签。"
    placeholder="搜索标签 (如: 治愈, 废萌)"
    :search="search"
  />
</template>
