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

const pending = ref(false)
const revert = (next: boolean) => {
  isFavorited.value = !next
  favoriteCount.value += next ? -1 : 1
}

const onChange = async (next: boolean) => {
  if (!id) {
    useAuthModal().open()
    revert(next)
    return
  }
  pending.value = true
  const result = await kunFetch(`/galgame/${props.galgameId}/favorite`, {
    method: 'PUT'
  })
  pending.value = false
  if (!result) {
    revert(next)
    return
  }
  useMessage(next ? 10526 : 10527, 'success')
}
</script>

<template>
  <KunTooltip text="收藏">
    <KunReaction
      v-model="isFavorited"
      v-model:count="favoriteCount"
      :disabled="pending"
      icon="lucide:heart"
      color="danger"
      label="收藏"
      @change="onChange"
    />
  </KunTooltip>
</template>
