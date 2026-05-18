<script setup lang="ts">
// Contributor management. Replaces the old read-only KunAvatarGroup:
// lists GET /galgame/:gid/contributors (richer than the detail payload's
// flat contributor[]) and lets the creator / admin remove one via
// DELETE /galgame/:gid/contributors/:uid. Backend verified: ProxyGet
// (router 124) + ProxyWriteWithToken DELETE (306); wiki gates
// creator/admin, we also gate the UI. See docs 03-relations.md.
const route = useRoute()
const gid = computed(() =>
  parseInt((route.params as { gid: string }).gid)
)

const galgame = inject<GalgameDetail>('galgame')
const { id: currentUserId, role } = usePersistUserStore()
const canManage = computed(
  () => galgame?.user.id === currentUserId || role >= 2
)

interface ContributorRow {
  id: number
  user_id: number
  created: string
  user: KunUser
}

const { data, status, refresh } = await useKunFetch<ContributorRow[]>(
  `/galgame/${gid.value}/contributors`,
  { lazy: true, method: 'GET' }
)

const removing = ref<number | null>(null)

const handleRemove = async (row: ContributorRow) => {
  const ok = await useComponentMessageStore().alert(
    `确定移除贡献者「${row.user.name}」吗?`,
    '该用户将不再计入本游戏的贡献者, 此操作不可撤销。'
  )
  if (!ok) return
  removing.value = row.user_id
  const res = await kunFetch(
    `/galgame/${gid.value}/contributors/${row.user_id}`,
    { method: 'DELETE' }
  )
  removing.value = null
  if (res !== null) {
    useMessage('已移除贡献者', 'success')
    refresh()
  }
}
</script>

<template>
  <div>
    <KunLoading v-if="status === 'pending'" />

    <KunNull
      v-else-if="!data || !data.length"
      description="暂无贡献者"
    />

    <div v-else class="flex flex-wrap gap-2">
      <div
        v-for="row in data"
        :key="row.id"
        class="border-default-200 flex items-center gap-2 rounded-full border py-1 pr-1 pl-2"
      >
        <KunUser :user="row.user" />
        <KunButton
          v-if="canManage"
          :is-icon-only="true"
          size="sm"
          rounded="full"
          variant="light"
          color="danger"
          :loading="removing === row.user_id"
          @click="handleRemove(row)"
        >
          <KunIcon name="lucide:x" />
        </KunButton>
      </div>
    </div>
  </div>
</template>
