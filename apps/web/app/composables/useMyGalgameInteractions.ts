// The current user's liked / favorited galgame ids, fetched once per session
// (client-side, logged-in only). The activity feed is cached SHARED across users
// (keyed by SFW setting, not per-user), so it can't carry the viewer's own
// like/favorite state — feed galgame cards hydrate their buttons from this
// instead. Shared via useState so every card reads one cache + one fetch.
export const useMyGalgameInteractions = () => {
  const { id } = usePersistUserStore()
  const liked = useState<number[]>('my-galgame-liked', () => [])
  const favorited = useState<number[]>('my-galgame-favorited', () => [])
  const loaded = useState<boolean>('my-galgame-interactions-loaded', () => false)

  const ensureLoaded = async () => {
    if (loaded.value || !id) return
    loaded.value = true // claim early so concurrent cards don't double-fetch
    const res = await kunFetch<{ liked: number[]; favorited: number[] }>(
      '/galgame/interactions/mine'
    )
    if (res) {
      liked.value = res.liked ?? []
      favorited.value = res.favorited ?? []
    } else {
      loaded.value = false // let a later mount retry
    }
  }

  const likedSet = computed(() => new Set(liked.value))
  const favoritedSet = computed(() => new Set(favorited.value))

  return {
    isLiked: (gid: number) => likedSet.value.has(gid),
    isFavorited: (gid: number) => favoritedSet.value.has(gid),
    ensureLoaded
  }
}
