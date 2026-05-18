<script setup lang="ts">
import type { UpdateGalgameTagPayload } from '~/components/galgame/types'
import {
  KUN_GALGAME_TAG_CATEGORY_MAP,
  type KUN_GALGAME_TAG_TYPE
} from '~/constants/galgameTag'

const { role } = usePersistUserStore()
const route = useRoute()
const tagId = computed(() => {
  return Number((route.params as { id: string }).id)
})

const pageData = reactive({
  page: 1,
  limit: 24,
  tagId: tagId.value
})

const showTagModal = ref(false)
const editingTag = ref<UpdateGalgameTagPayload>({} as UpdateGalgameTagPayload)

const { data, status } = await useKunFetch(`/galgame-tag/${tagId.value}`, {
  method: 'GET',
  query: pageData
})

const openEditTagModal = () => {
  if (!data.value) {
    return
  }
  const res = data.value
  editingTag.value = {
    name: res.name,
    tagId: res.id,
    description: res.description,
    category: res.category as (typeof KUN_GALGAME_TAG_TYPE)[number],
    alias: res.alias
  } satisfies UpdateGalgameTagPayload
  showTagModal.value = true
}

const handleUpdateTag = async (data: UpdateGalgameTagPayload) => {
  const result = await kunFetch(`/galgame-tag`, {
    method: 'PUT',
    body: data
  })

  if (result) {
    useMessage('重新编辑成功', 'success')
  }
}

// DELETE /galgame-tag/:id → wiki DELETE /tag/:id (router 316; cascades
// aliases + relations + Meilisearch). admin/moderator only — wiki gates
// it; UI gated at role>=2 (wiki semantics; §15.2 forbids narrowing).
const isDeleting = ref(false)
const handleDeleteTag = async () => {
  const ok = await useComponentMessageStore().alert(
    `确定删除标签「${data.value?.name}」吗?`,
    '将级联移除该标签的别名与所有 Galgame 关联, 不可撤销。'
  )
  if (!ok) return
  isDeleting.value = true
  const res = await kunFetch(`/galgame-tag/${tagId.value}`, {
    method: 'DELETE'
  })
  isDeleting.value = false
  if (res !== null) {
    useMessage('标签已删除', 'success')
    await navigateTo('/galgame-tag')
  }
}
</script>

<template>
  <KunCard
    :is-transparent="false"
    :is-hoverable="false"
    :is-pressable="false"
    content-class="space-y-6"
    v-if="data"
  >
    <KunHeader
      :name="`含有标签 ${data.name} 的 Galgame`"
      :description="data.description"
    >
      <template #endContent>
        <div class="space-y-3">
          <p class="text-default-500">
            默认仅显示了 SFW 的 Galgame, 查看 NSFW Galgame 请在设置面板打开 NSFW
            开关。如果有数据错误请
            <KunLink to="/doc/contact"> 联系我们 </KunLink>。
          </p>

          <div class="text-default-500">
            标签类别
            <KunBadge
              :color="
                data.category === 'content'
                  ? 'primary'
                  : data.category === 'sexual'
                    ? 'danger'
                    : 'success'
              "
            >
              {{ KUN_GALGAME_TAG_CATEGORY_MAP[data.category] }}
            </KunBadge>
          </div>
          <div
            v-if="data.alias.length"
            class="text-default-500 flex flex-wrap gap-2"
          >
            别名
            <KunBadge
              color="primary"
              v-for="(a, index) in data.alias"
              :key="index"
            >
              {{ a }}
            </KunBadge>
          </div>
          <div v-if="role >= 2" class="flex justify-end gap-2">
            <KunButton @click="openEditTagModal">编辑标签</KunButton>
            <KunButton
              variant="flat"
              color="danger"
              :loading="isDeleting"
              @click="handleDeleteTag"
            >
              删除标签
            </KunButton>
          </div>
        </div>
      </template>
    </KunHeader>

    <GalgameTagModal
      v-model="showTagModal"
      :initial-data="editingTag"
      @submit="handleUpdateTag"
    />

    <GalgameCard
      :is-transparent="true"
      v-if="data.galgame.length"
      :galgames="data.galgame"
    />

    <KunPagination
      v-if="data.galgameCount > pageData.limit"
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.galgameCount / pageData.limit)"
      :is-loading="status === 'pending'"
    />

    <KunNull
      v-if="!data.galgameCount"
      :description="`${data.name} 标签下暂无 Galgame`"
    />
  </KunCard>
</template>
