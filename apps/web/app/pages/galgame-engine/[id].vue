<script setup lang="ts">
import type { UpdateGalgameEnginePayload } from '~/components/galgame/types'

const { role } = usePersistUserStore()
const route = useRoute()
const engineId = computed(() => {
  return Number((route.params as { id: string }).id)
})

// Shared store with /galgame-tag/[id].vue + /galgame-official/[id].vue
// so the global galgame filter Nav (sort/type/language/platform)
// drives the embedded galgame list here too. Previously a local
// `pageData` snapshot meant the entity_handler's sortField/sortOrder
// rename helper had no params to translate on this page.
const { page, limit, type, language, platform, sortField, sortOrder } =
  useGalgameFilters()

const showEngineModal = ref(false)
const editingEngine = ref<UpdateGalgameEnginePayload>(
  {} as UpdateGalgameEnginePayload
)

const { data, status } = await useKunFetch<GalgameEngineDetail>(
  `/galgame-engine/${engineId.value}`,
  {
    method: 'GET',
    query: {
      page,
      limit,
      type,
      language,
      platform,
      sortField,
      sortOrder,
      engineId
    }
  }
)

const openEditEngineModal = () => {
  if (!data.value) {
    return
  }
  const res = data.value
  editingEngine.value = {
    engineId: res.id,
    name: res.name,
    description: res.description,
    alias: res.alias
  } satisfies UpdateGalgameEnginePayload
  showEngineModal.value = true
}

const handleUpdateEngine = async (data: UpdateGalgameEnginePayload) => {
  const result = await kunFetch(`/galgame-engine`, {
    method: 'PUT',
    body: data
  })

  if (result) {
    useMessage('重新编辑成功', 'success')
  }
}

// Two-stage safe delete (docs 04-taxonomy / 00-handbook): plain DELETE
// is rejected while still referenced (wiki toasts the count); only after
// an explicit second confirm do we retry ?force=true to purge relations
// + hard delete. admin/moderator only — wiki gates; UI role>=2 (§15.2).
const isDeleting = ref(false)
const handleDeleteEngine = async () => {
  const ok = await useComponentMessageStore().alert(
    `确定删除引擎「${data.value?.name}」吗?`,
    '若该引擎未被任何 Galgame 引用将直接删除; 仍被引用时会先提示。'
  )
  if (!ok) return
  isDeleting.value = true
  const res = await kunFetch(`/galgame-engine/${engineId.value}`, {
    method: 'DELETE'
  })
  if (res !== null) {
    isDeleting.value = false
    useMessage('引擎已删除', 'success')
    await navigateTo('/galgame-engine')
    return
  }
  isDeleting.value = false
  const force = await useComponentMessageStore().alert(
    '该引擎仍被 Galgame 引用, 删除已被拒绝',
    '强制删除会先清除该引擎在所有 Galgame 上的关联, 再硬删除该引擎, 不可撤销。确定强制删除吗?'
  )
  if (!force) return
  isDeleting.value = true
  const forced = await kunFetch(`/galgame-engine/${engineId.value}`, {
    method: 'DELETE',
    query: { force: true }
  })
  isDeleting.value = false
  if (forced !== null) {
    useMessage('引擎已强制删除', 'success')
    await navigateTo('/galgame-engine')
  }
}

if (data.value) {
  useKunSeoMeta({
    title: `${data.value.name} 引擎`,
    description: `查看所有使用 ${data.value.name} 引擎制作的 Galgame`
  })
} else {
  useKunDisableSeo('未找到 Galgame 引擎')
}
</script>

<template>
  <KunCard
    :is-transparent="false"
    :is-hoverable="false"
    content-class="space-y-6"
    v-if="data"
  >
    <KunHeader
      :name="`${data.name} 引擎制作的 Galgame`"
      :description="data.description"
    >
      <template #endContent>
        <div class="space-y-3">
          <p class="text-default-500">
            默认仅显示了 SFW 的 Galgame, 查看 NSFW Galgame 请在设置面板打开 NSFW
            开关。如果有数据错误请
            <KunLink to="/doc/contact"> 联系我们 </KunLink>。
          </p>

          <div
            v-if="data.alias.length"
            class="text-default-500 flex flex-wrap gap-2"
          >
            别名
            <KunChip
              color="primary"
              v-for="(a, index) in data.alias"
              :key="index"
            >
              {{ a }}
            </KunChip>
          </div>
          <div class="flex flex-wrap justify-end gap-2">
            <GalgameRevisionModal
              entity="engine"
              :id="engineId"
              :entity-label="`引擎「${data.name}」`"
              :can-revert="role >= 2"
            />
            <template v-if="role >= 2">
              <KunButton @click="openEditEngineModal">编辑引擎</KunButton>
              <KunButton
                variant="flat"
                color="danger"
                :loading="isDeleting"
                @click="handleDeleteEngine"
              >
                删除引擎
              </KunButton>
            </template>
          </div>
        </div>
      </template>
    </KunHeader>

    <GalgameEngineModal
      v-model="showEngineModal"
      :initial-data="editingEngine"
      @submit="handleUpdateEngine"
    />

    <GalgameCard
      :is-transparent="true"
      v-if="data.galgame.length"
      :galgames="data.galgame"
    />

    <KunPagination
      v-if="data.galgameCount > limit"
      v-model:current-page="page"
      :total-page="Math.ceil(data.galgameCount / limit)"
      :is-loading="status === 'pending'"
    />

    <KunNull
      v-if="!data.galgameCount"
      :description="`${data.name} 引擎下暂无 Galgame`"
    />
  </KunCard>
</template>
