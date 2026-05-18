<script setup lang="ts">
// Revision history + per-revision diff (GET /galgame/:gid/revisions/
// :rev/diff → { changed_keys, old, new }, rendered by the shared
// GalgameSnapshotDiff) + revert (POST /galgame/:gid/revert {revision},
// creator/admin only — wiki gates it; we also gate the UI). Backend
// paths verified: /history/all → wiki /revisions (router 164),
// /revisions/:rev/diff ProxyGet (119), /revert ProxyWriteWithToken
// (299). See docs/galgame_wiki/02-revisions-and-prs.md.
const KUN_REVISION_ACTION_MAP: Record<string, string> = {
  created: '创建',
  updated: '编辑',
  merged: '合并 PR',
  reverted: '回滚',
  declined: '拒绝 PR'
}

const route = useRoute()
const gid = computed(() => {
  return parseInt((route.params as { gid: string }).gid)
})

const galgame = inject<GalgameDetail>('galgame')
const { id: currentUserId, role } = usePersistUserStore()
const canRevert = computed(
  () => galgame?.user.id === currentUserId || role >= 2
)

const pageData = reactive({
  page: 1,
  limit: 10,
  galgameId: gid.value
})

const { data, status, refresh } = await useKunFetch<{
  items: GalgameRevision[]
  total: number
}>(`/galgame/${gid.value}/history/all`, {
  lazy: true,
  method: 'GET',
  query: pageData
})

interface RevDiff {
  changed_keys: Record<string, boolean>
  old: Record<string, unknown>
  new: Record<string, unknown>
}

const openRev = ref<number | null>(null)
const diffCache = reactive<Record<number, RevDiff>>({})
const diffLoading = ref<number | null>(null)
const reverting = ref<number | null>(null)

const toggleDiff = async (rev: number) => {
  if (openRev.value === rev) {
    openRev.value = null
    return
  }
  openRev.value = rev
  if (diffCache[rev]) return
  diffLoading.value = rev
  const res = await kunFetch<RevDiff>(
    `/galgame/${gid.value}/revisions/${rev}/diff`,
    { method: 'GET' }
  )
  diffLoading.value = null
  if (res) diffCache[rev] = res
}

const handleRevert = async (rev: number) => {
  const ok = await useComponentMessageStore().alert(
    `确定回滚到版本 #${rev} 吗?`,
    '回滚会基于该版本创建一条新的版本记录, 不会删除任何历史。'
  )
  if (!ok) return
  reverting.value = rev
  const res = await kunFetch(`/galgame/${gid.value}/revert`, {
    method: 'POST',
    body: { revision: rev }
  })
  reverting.value = null
  if (res) {
    useMessage('回滚成功', 'success')
    openRev.value = null
    refresh()
  }
}
</script>

<template>
  <div class="flex flex-col space-y-3" v-if="data">
    <KunHeader
      name="版本历史"
      description="这里记录了这个 Galgame 项目发生的所有更改历史"
      scale="h3"
    />

    <KunLoading v-if="status === 'pending'" />

    <div
      v-for="(rev, index) in data.items"
      :key="index"
      class="border-default-200 space-y-2 rounded-lg border p-3"
    >
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2 text-sm">
          <KunAvatar :user="rev.user" />
          <div class="space-y-1">
            <div class="flex flex-wrap items-center gap-2">
              <span>{{ rev.user.name }}</span>
              <KunBadge size="sm">
                {{ KUN_REVISION_ACTION_MAP[rev.action] || rev.action }}
              </KunBadge>
              <span class="text-default-400 text-xs">#{{ rev.revision }}</span>
              <span v-if="rev.isMinor" class="text-default-400 text-xs">
                (小修改)
              </span>
              <span class="text-default-500">
                {{ formatTimeDifference(rev.created) }}
              </span>
            </div>
            <div class="text-default-500" v-if="rev.note">
              {{ rev.note }}
            </div>
          </div>
        </div>

        <div class="flex shrink-0 items-center gap-1">
          <KunButton
            size="sm"
            variant="flat"
            @click="toggleDiff(rev.revision)"
            :loading="diffLoading === rev.revision"
          >
            {{ openRev === rev.revision ? '收起' : '差异' }}
          </KunButton>
          <KunButton
            v-if="canRevert"
            size="sm"
            variant="flat"
            color="warning"
            :loading="reverting === rev.revision"
            @click="handleRevert(rev.revision)"
          >
            回滚到此版本
          </KunButton>
        </div>
      </div>

      <div
        v-if="openRev === rev.revision && diffCache[rev.revision]"
        class="border-default-200 border-t pt-2"
      >
        <GalgameSnapshotDiff
          :changed-keys="diffCache[rev.revision]!.changed_keys"
          :old-snap="diffCache[rev.revision]!.old"
          :new-snap="diffCache[rev.revision]!.new"
        />
      </div>
    </div>

    <KunPagination
      v-if="data.total >= pageData.limit"
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.total / pageData.limit)"
      :is-loading="status === 'pending'"
    />
  </div>
</template>
