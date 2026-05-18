<script setup lang="ts">
// Tag picker for the PR edit form. Network search via
// /galgame-tag/search; "没有就新建" via the create modal (POST
// /galgame-tag, any logged-in user). Binds full GalgameTagItem objects
// into the temp PR store; Footer derives tag_ids[] at submit.
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

const baseRef = ref<{ addCreated: (e: GalgameTagItem) => void } | null>(null)
const modalOpen = ref(false)
const newName = ref('')

const onCreate = (q: string) => {
  newName.value = q
  modalOpen.value = true
}
const onCreated = (entity: { id: number; name: string }) => {
  baseRef.value?.addCreated(entity as GalgameTagItem)
}
</script>

<template>
  <div class="contents">
    <EditGalgamePrEntitySelectorBase
      ref="baseRef"
      v-model="model"
      title="标签"
      description="输入关键字搜索 Wiki 已有标签, 点击添加; 没有就新建。提交后整组替换该 Galgame 的标签。"
      placeholder="搜索标签 (如: 治愈, 废萌)"
      :search="search"
      creatable
      @create="onCreate"
    />
    <EditGalgamePrEntityCreateModal
      v-model="modalOpen"
      type="tag"
      :initial-name="newName"
      @created="onCreated"
    />
  </div>
</template>
