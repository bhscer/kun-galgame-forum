import type { InjectionKey, Ref, ComputedRef } from 'vue'

// Shared reaction state for a single topic / reply. The owning component
// (Master.vue / Reply.vue) creates it and provides it via reactionsKey so the
// chips (TopicReactionBar) and the trigger button (TopicReactionTrigger) — which
// may live apart, e.g. the trigger in the desktop footer — stay in sync.
export interface ReactionsState {
  list: Ref<KunReaction[]>
  mineKeys: ComputedRef<string[]>
  toggle: (key: string) => Promise<void>
}

export const reactionsKey: InjectionKey<ReactionsState> = Symbol('reactions')

interface UseReactionsOptions {
  topicId?: number
  replyId?: number
  targetUserId: number
  reactions: KunReaction[]
}

// Avatars are shown for reactions with fewer than this many reactors (matches
// the backend hydration threshold); above it we show emoji + count.
const AVATAR_THRESHOLD = 5

export const useReactions = (opts: UseReactionsOptions): ReactionsState => {
  const { id, name, avatar } = usePersistUserStore()

  const clone = (rs: KunReaction[]): KunReaction[] =>
    rs.map((r) => ({ ...r, reactors: r.reactors ? [...r.reactors] : undefined }))

  const list = ref<KunReaction[]>(opts.reactions ? clone(opts.reactions) : [])
  const inflight = new Set<string>()

  const mineKeys = computed(() =>
    list.value.filter((r) => r.mine).map((r) => r.reaction)
  )

  const post = (reaction: string) =>
    opts.topicId
      ? kunFetch(`/topic/${opts.topicId}/reaction`, {
          method: 'PUT',
          body: { reaction }
        })
      : kunFetch(`/topic/0/reply/reaction`, {
          method: 'PUT',
          body: { replyId: opts.replyId, reaction }
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
    if (!id) {
      useAuthModal().open()
      return
    }
    if (key === 'like' && id === opts.targetUserId) {
      useMessage(10236, 'warn')
      return
    }
    if (inflight.has(key)) return
    inflight.add(key)

    const snapshot = clone(list.value)
    applyOptimistic(key)

    const ok = await post(key)
    if (!ok) list.value = snapshot
    inflight.delete(key)
  }

  return { list, mineKeys, toggle }
}
