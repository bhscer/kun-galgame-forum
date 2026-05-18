<script setup lang="ts">
// Engine picker. No /galgame-engine/search endpoint — /galgame-engine
// returns the full list; lazily fetched once, filtered locally over
// name + alias. "没有就新建" via POST /galgame-engine.
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

const baseRef = ref<{ addCreated: (e: GalgameEngineItem) => void } | null>(
  null
)
const modalOpen = ref(false)
const newName = ref('')

const onCreate = (q: string) => {
  newName.value = q
  modalOpen.value = true
}
const onCreated = (entity: { id: number; name: string }) => {
  // Drop the stale cached list so a later search re-fetches and the new
  // engine shows up there too.
  all.value = null
  baseRef.value?.addCreated(entity as GalgameEngineItem)
}
</script>

<template>
  <div class="contents">
    <EditGalgamePrEntitySelectorBase
      ref="baseRef"
      v-model="model"
      title="游戏引擎"
      description="搜索 Wiki 已有引擎并添加, 没有就新建; 提交后整组替换。"
      placeholder="搜索引擎 (如: KiriKiri, Unity)"
      :search="search"
      creatable
      @create="onCreate"
    />
    <EditGalgamePrEntityCreateModal
      v-model="modalOpen"
      type="engine"
      :initial-name="newName"
      @created="onCreated"
    />
  </div>
</template>
