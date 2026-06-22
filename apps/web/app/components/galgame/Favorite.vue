<script setup lang="ts">
const props = defineProps<{
  galgameId: number
  targetUserId: number
  favoriteCount: number
  isFavorited: boolean
}>()

const { id } = usePersistUserStore()
const isFavorited = ref(props.isFavorited)
const favoriteCount = ref(props.favoriteCount)

// The feed hydrates is-favorited asynchronously (see useMyGalgameInteractions),
// so reflect a late-arriving initial state. Harmless on the detail page where
// the prop is already settled.
watch(
  () => props.isFavorited,
  (value) => (isFavorited.value = value)
)

const toggleFavoriteGalgame = async () => {
  const result = await kunFetch(`/galgame/${props.galgameId}/favorite`, {
    method: 'PUT'
  })

  if (result) {
    favoriteCount.value += isFavorited.value ? -1 : 1

    if (!isFavorited.value) {
      useMessage(10526, 'success')
    } else {
      useMessage(10527, 'success')
    }

    isFavorited.value = !isFavorited.value
  }
}

const handleClickFavoriteThrottled = throttle(toggleFavoriteGalgame, 1007, () =>
  useMessage(10528, 'warn')
)

const handleClickFavorite = () => {
  if (!id) {
    useAuthModal().open()
    return
  }
  handleClickFavoriteThrottled()
}
</script>

<template>
  <KunTooltip text="收藏">
    <KunButton
      :variant="isFavorited ? 'flat' : 'light'"
      :color="isFavorited ? 'secondary' : 'default'"
      :size="favoriteCount ? 'md' : 'lg'"
      class-name="gap-1"
      @click="handleClickFavorite"
    >
      <KunIcon name="lucide:heart" />
      <span v-if="favoriteCount">{{ favoriteCount }}</span>
    </KunButton>
  </KunTooltip>
</template>
