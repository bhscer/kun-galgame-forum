<script setup lang="ts">
// 萌萌点明细 modal. Opened from the avatar menu (UserInfo.vue) via the
// temp-store flag `showKUNGalgameMoemoepointLog`, and rendered at the app.vue
// root (a stable, NON-scoped node) — NOT inside the avatar popover, whose
// content is v-if'd and unmounts the instant the user clicks into the modal
// (which would tear the modal down with it). app.vue (vs the <style scoped>
// avatar bar) also keeps Vue from warning about the scope id it can't stamp
// onto this component's <KunModal> teleport root.
//
// The ledger is the UNIFIED moemoepoint history from OAuth: every earn / spend
// across all sites (鲲 Galgame / 补丁 / TouchGal / 贴纸 …) shows up here, since
// the balance is a single source of truth. Cursor pagination by `before_id`.
interface MoemoepointLogEntry {
  id: number
  delta: number
  reason: string
  source_app: string
  ref: string
  created_at: string
}

const PAGE_SIZE = 20

const { showKUNGalgameMoemoepointLog: isOpen } = storeToRefs(
  useTempSettingStore()
)
const { moemoepoint } = storeToRefs(usePersistUserStore())

// reason is OAuth's small stable enum (06-moemoepoint.md §2). admin_* /
// migration are server-side only but can still appear in a user's history.
const REASON_META: Record<string, { label: string; icon: string }> = {
  daily_checkin: { label: '每日签到', icon: 'lucide:calendar-check' },
  liked: { label: '内容被点赞', icon: 'lucide:heart' },
  content_approved: { label: '内容被采纳', icon: 'lucide:circle-check-big' },
  content_removed: { label: '内容被移除', icon: 'lucide:circle-x' },
  admin_grant: { label: '管理员发放', icon: 'lucide:gift' },
  admin_deduct: { label: '管理员扣除', icon: 'lucide:gavel' },
  migration: { label: '初始迁移', icon: 'lucide:database' },
  register_gift: { label: '注册礼物', icon: 'lucide:party-popper' }
}

// source_app is whatever OAuth derives from the calling client. Today OAuth
// returns the raw client_id (an opaque 32-hex hash), so a friendly name can't
// be resolved client-side for cross-site entries — only OAuth owns the
// client→app registry. We map known readable slugs (in case OAuth starts
// sending them) and HIDE anything opaque rather than print a hash.
const SOURCE_LABEL: Record<string, string> = {
  kungal: '鲲 Galgame',
  moyu: '鲲补丁',
  patch: '鲲补丁',
  touchgal: 'TouchGal',
  sticker: '贴纸小铺',
  stickers: '贴纸小铺',
  oauth: '账号中心'
}

const REF_KIND_LABEL: Record<string, string> = {
  galgame: 'Galgame',
  galgame_pr: 'Galgame 修订',
  galgame_comment: '游戏评论',
  galgame_rating: '游戏评分',
  galgame_resource: '游戏资源',
  toolset: '工具集',
  toolset_resource: '工具资源',
  topic: '话题',
  topic_comment: '话题评论',
  topic_reply: '回复'
}

const reasonMeta = (reason: string) =>
  REASON_META[reason] ?? {
    label: reason || '萌萌点变动',
    icon: 'lucide:lollipop'
  }

// OAuth keeps a tiny generic reason enum; the concrete behavior is carried by
// the ref-kind (06-moemoepoint.md §2). So the PRIMARY label is derived from
// (reason, ref-kind): a precise override first, else a composed "<内容><动作>"
// (e.g. 游戏资源 + 被采纳), else the bare reason label. This is why a single
// `content_approved` row reads as 通过审核 / 修订被合并 / 资源被采纳 / 评论被采纳…
const REASON_ACTION: Record<string, string> = {
  content_approved: '被采纳',
  liked: '被点赞',
  content_removed: '被移除'
}

const BEHAVIOR_LABEL: Record<string, string> = {
  'content_approved:galgame': 'Galgame 通过审核',
  'content_approved:galgame_pr': 'Galgame 修订被合并',
  'content_approved:galgame_resource': '发布的资源被采纳',
  'content_approved:galgame_comment': '游戏评论被采纳',
  'content_approved:topic': '发布的话题被采纳',
  'content_approved:topic_reply': '回复被采纳',
  'content_approved:topic_comment': '话题评论被采纳',
  // 推话题: a distinct ref-kind so the cost reads as 推话题消耗 (not 话题被移除) and
  // the owner's credit as 话题被推荐 (not 被采纳). Reasons stay OAuth's enum.
  'content_removed:topic_upvote': '推话题消耗',
  'content_approved:topic_upvote': '话题被推荐'
}

const refKindOf = (ref: string) => ref.split(':')[0] ?? ''

const behaviorLabel = (entry: MoemoepointLogEntry): string => {
  const kind = refKindOf(entry.ref)
  const specific = BEHAVIOR_LABEL[`${entry.reason}:${kind}`]
  if (specific) return specific
  const kindLabel = REF_KIND_LABEL[kind]
  const action = REASON_ACTION[entry.reason]
  if (kindLabel && action) return `${kindLabel}${action}`
  return reasonMeta(entry.reason).label
}

const isOpaqueId = (value: string) => /^[0-9a-f]{16,}$/i.test(value)

const sourceLabel = (app: string): string => {
  if (!app) return ''
  const slug = app.replace(/-backend$/, '')
  if (SOURCE_LABEL[slug]) return SOURCE_LABEL[slug]
  return isOpaqueId(slug) ? '' : slug
}

// The ref-kind is now shown in the primary label, so the sub-line only adds
// the entity id (#123) for disambiguation — empty when the ref has no id.
const refId = (refValue: string): string => {
  const id = refValue.split(':')[1]
  return id ? `#${id}` : ''
}

// One muted line under the behavior: "source · #id · time", omitting empties.
const entryMeta = (entry: MoemoepointLogEntry): string =>
  [
    sourceLabel(entry.source_app),
    refId(entry.ref),
    formatTimeDifference(entry.created_at)
  ]
    .filter(Boolean)
    .join(' · ')

const entries = ref<MoemoepointLogEntry[]>([])
const status = ref<'idle' | 'loading' | 'loadingMore' | 'error'>('idle')
const hasMore = ref(true)

const fetchPage = async (more = false) => {
  if (more && (!hasMore.value || status.value === 'loadingMore')) return
  status.value = more ? 'loadingMore' : 'loading'

  const beforeId =
    more && entries.value.length
      ? entries.value[entries.value.length - 1]!.id
      : 0

  const page = await kunFetch<{
    items: MoemoepointLogEntry[]
    has_more: boolean
  }>('/user/moemoepoint/log', {
    query: { limit: PAGE_SIZE, before_id: beforeId }
  })

  if (page === null) {
    status.value = 'error'
    return
  }

  entries.value = more ? [...entries.value, ...page.items] : page.items
  hasMore.value = page.has_more
  status.value = 'idle'
}

// Refetch on each open so a just-earned point (e.g. fresh check-in) shows.
watch(isOpen, (open) => {
  if (!open) return
  entries.value = []
  hasMore.value = true
  fetchPage(false)
})
</script>

<template>
  <KunModal v-model="isOpen" inner-class-name="max-w-lg w-full">
    <div class="flex max-h-[75dvh] flex-col gap-3 p-1">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <KunIcon class="text-secondary text-2xl" name="lucide:lollipop" />
          <span class="text-lg font-bold">萌萌点明细</span>
        </div>
        <p class="flex items-center gap-1 font-bold">
          <span class="text-default-500 text-sm font-normal">当前</span>
          <span class="text-secondary">{{ moemoepoint }}</span>
        </p>
      </div>

      <p class="text-default-500 text-xs">
        这里汇总了你在鲲 Galgame 全站(及关联站点)的萌萌点收支记录
      </p>

      <KunLoading v-if="status === 'loading'" />

      <KunNull
        v-else-if="status === 'error'"
        description="加载失败, 请稍后再试"
      />

      <KunNull v-else-if="!entries.length" description="暂无萌萌点记录" />

      <div v-else class="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto">
        <div
          v-for="entry in entries"
          :key="entry.id"
          class="hover:bg-default-100 flex items-center gap-3 rounded-lg p-2 transition-colors"
        >
          <KunIcon
            class="text-default-500 shrink-0 text-xl"
            :name="reasonMeta(entry.reason).icon"
          />
          <div class="flex min-w-0 grow flex-col">
            <span class="truncate text-sm font-medium">
              {{ behaviorLabel(entry) }}
            </span>
            <span class="text-default-400 truncate text-xs">
              {{ entryMeta(entry) }}
            </span>
          </div>
          <span
            class="shrink-0 text-sm font-bold tabular-nums"
            :class="entry.delta >= 0 ? 'text-success-600' : 'text-danger-500'"
          >
            {{ entry.delta >= 0 ? '+' : '' }}{{ entry.delta }}
          </span>
        </div>

        <KunButton
          v-if="hasMore"
          variant="light"
          class-name="mt-1"
          :loading="status === 'loadingMore'"
          @click="fetchPage(true)"
        >
          加载更多
        </KunButton>
      </div>
    </div>
  </KunModal>
</template>
