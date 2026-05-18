<script setup lang="ts">
// Engine picker. Unlike tag/official there is NO /galgame-engine/search
// endpoint — /galgame-engine returns the full list. We lazily fetch it
// once on first interaction (cache in a ref) and filter locally over
// name + alias, so the dropdown UX stays identical to the others.
const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const model = computed<GalgameEngineItem[]>({
  get: () => galgamePR.value[0]?.engines ?? [],
  set: (v) => {
    if (galgamePR.value[0]) galgamePR.value[0].engines = v
  }
})

const all = ref<GalgameEngineItem[] | null>(null)

const search = async (q: string): Promise<GalgameEngineItem[]> => {
  if (all.value === null) {
    all.value =
      (await kunFetch<GalgameEngineItem[]>('/galgame-engine', {
        method: 'GET'
      })) ?? []
  }
  const kw = q.toLowerCase()
  return all.value
    .filter(
      (e) =>
        e.name.toLowerCase().includes(kw) ||
        e.alias?.some((a) => a.toLowerCase().includes(kw))
    )
    .slice(0, 50)
}
</script>

<template>
  <EditGalgamePrEntitySelectorBase
    v-model="model"
    title="游戏引擎"
    description="搜索 Wiki 已有引擎并添加, 提交后整组替换。"
    placeholder="搜索引擎 (如: KiriKiri, Unity)"
    :search="search"
  />
</template>
