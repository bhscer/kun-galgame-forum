<script setup lang="ts">
// Telegram-style reaction bar below a topic / reply. Shows each reaction as a
// chip — reactor avatars when count < 5, else emoji + count — tap to toggle; a
// trailing "+" opens the picker. Self-contained: owns a local copy of the
// reactions and updates it optimistically (count + mine + like⇄dislike
// exclusion), reverting on a failed request. Dual-mode like the like button:
// pass topicId OR replyId.
import { KUN_REACTION_EMOJI, reactionAsset } from '~/constants/reaction'

const props = defineProps<{
  topicId?: number
  replyId?: number
  targetUserId: number
  reactions: KunReaction[]
}>()

const AVATAR_THRESHOLD = 5

const { id, name, avatar } = usePersistUserStore()

const clone = (rs: KunReaction[]): KunReaction[] =>
  rs.map((r) => ({ ...r, reactors: r.reactors ? [...r.reactors] : undefined }))

const list = ref<KunReaction[]>(props.reactions ? clone(props.reactions) : [])
watch(
  () => props.reactions,
  (v) => (list.value = v ? clone(v) : [])
)

const inflight = ref(new Set<string>())
const picker = ref<{ close: () => void } | null>(null)

const mineKeys = computed(() =>
  list.value.filter((r) => r.mine).map((r) => r.reaction)
)

const showAvatars = (r: KunReaction) =>
  r.count < AVATAR_THRESHOLD && !!r.reactors?.length

const post = (reaction: string) =>
  props.topicId
    ? kunFetch(`/topic/${props.topicId}/reaction`, {
        method: 'PUT',
        body: { reaction }
      })
    : kunFetch(`/topic/0/reply/reaction`, {
        method: 'PUT',
        body: { replyId: props.replyId, reaction }
      })

const removeMine = (idx: number) => {
  const r = list.value[idx]!
  r.count--
  r.mine = false
  if (r.reactors) r.reactors = r.reactors.filter((u) => u.id !== id)
  if (r.count <= 0) list.value.splice(idx, 1)
}

const applyOptimistic = (key: string) => {
  const idx = list.value.findIndex((r) => r.reaction === key)
  if (idx >= 0 && list.value[idx]!.mine) {
    removeMine(idx)
  } else if (idx >= 0) {
    const r = list.value[idx]!
    r.count++
    r.mine = true
    if (r.count < AVATAR_THRESHOLD) {
      ;(r.reactors ??= []).push({ id, name, avatar })
    }
  } else {
    list.value.push({
      reaction: key,
      count: 1,
      mine: true,
      reactors: [{ id, name, avatar }]
    })
  }
  // like ⇄ dislike are mutually exclusive — drop the opposite if it was mine.
  const opposite =
    key === 'like' ? 'dislike' : key === 'dislike' ? 'like' : null
  if (opposite) {
    const j = list.value.findIndex((r) => r.reaction === opposite)
    if (j >= 0 && list.value[j]!.mine) removeMine(j)
  }
}

const toggle = async (key: string) => {
  picker.value?.close()
  if (!id) {
    useAuthModal().open()
    return
  }
  if (key === 'like' && id === props.targetUserId) {
    useMessage(10236, 'warn')
    return
  }
  if (inflight.value.has(key)) return
  inflight.value.add(key)

  const snapshot = clone(list.value)
  applyOptimistic(key)

  const ok = await post(key)
  if (!ok) list.value = snapshot
  inflight.value.delete(key)
}
</script>

<template>
  <div v-if="list.length || id" class="flex flex-wrap items-center gap-1.5">
    <button
      v-for="r in list"
      :key="r.reaction"
      type="button"
      :class="
        cn(
          'flex items-center gap-1 rounded-full border px-2 py-0.5 text-sm transition-colors',
          r.mine
            ? 'border-primary bg-primary/10 text-primary'
            : 'border-default-200 text-default-600 hover:bg-default-100'
        )
      "
      @click="toggle(r.reaction)"
    >
      <img
        :src="reactionAsset(r.reaction)"
        :alt="KUN_REACTION_EMOJI[r.reaction] ?? r.reaction"
        class="size-5 shrink-0 max-w-none"
        loading="lazy"
      />
      <span v-if="showAvatars(r)" class="flex -space-x-1.5">
        <KunAvatar
          v-for="u in r.reactors!.slice(0, 4)"
          :key="u.id"
          :user="u"
          size="sm"
          :is-navigation="false"
        />
      </span>
      <span v-else class="tabular-nums">{{ formatNumber(r.count) }}</span>
    </button>

    <KunPopover ref="picker" position="top-start">
      <template #trigger>
        <button
          type="button"
          class="border-default-200 text-default-500 hover:bg-default-100 flex items-center rounded-full border px-2 py-1 transition-colors"
        >
          <KunIcon name="lucide:smile-plus" class="size-4" />
        </button>
      </template>

      <TopicReactionPicker :mine-keys="mineKeys" @select="toggle" />
    </KunPopover>
  </div>
</template>
