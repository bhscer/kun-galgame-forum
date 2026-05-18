<script setup lang="ts">
// Developer/publisher (会社) picker. /galgame-official/search expects a
// space-tokenized `q` array (same as components/galgame/official/
// Container.vue). "没有就新建" via POST /galgame-official. Binds full
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

const baseRef = ref<{
  addCreated: (e: GalgameOfficialItem) => void
} | null>(null)
const modalOpen = ref(false)
const newName = ref('')

const onCreate = (q: string) => {
  newName.value = q
  modalOpen.value = true
}
const onCreated = (entity: { id: number; name: string }) => {
  baseRef.value?.addCreated(entity as GalgameOfficialItem)
}
</script>

<template>
  <div class="contents">
    <EditGalgamePrEntitySelectorBase
      ref="baseRef"
      v-model="model"
      title="会社"
      description="开发商 / 发行商。搜索 Wiki 已有会社并添加, 没有就新建; 提交后整组替换。"
      placeholder="搜索会社 (空格分隔多个关键字)"
      :search="search"
      creatable
      @create="onCreate"
    />
    <EditGalgamePrEntityCreateModal
      v-model="modalOpen"
      type="official"
      :initial-name="newName"
      @created="onCreated"
    />
  </div>
</template>
