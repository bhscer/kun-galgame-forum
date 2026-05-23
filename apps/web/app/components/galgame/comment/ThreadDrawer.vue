<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

// Side drawer that renders an entire comment thread from its root.
//
// Triggered when a deep reply inside the inline list hits the visible-
// depth cap. The inline list caps recursion at 3 layers for readability;
// the drawer is depth-unbounded so the user can see the full
// conversation in a focused side panel without the rest of the page
// fighting for attention.
//
// Re-fetches the thread on open (instead of consuming the already-
// loaded subtree) so the drawer reflects new replies posted in another
// tab / window between page load and click.
const props = defineProps<{
  galgameId: number
  rootCommentId: number | null
  refresh: () => void
}>()

const emit = defineEmits<{
  'update:rootCommentId': [value: number | null]
}>()

const isOpen = computed({
  get: () => props.rootCommentId !== null,
  set: (value) => {
    if (!value) emit('update:rootCommentId', null)
  }
})

const thread = ref<SerializeObject<GalgameComment> | null>(null)
const isLoading = ref(false)

const loadThread = async (rootId: number) => {
  isLoading.value = true
  thread.value = null
  try {
    const data = await kunFetch<SerializeObject<GalgameComment>>(
      `/galgame/${props.galgameId}/comment/thread/${rootId}`,
      { method: 'GET' }
    )
    if (data) thread.value = data
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.rootCommentId,
  (rootId) => {
    if (rootId !== null) loadThread(rootId)
    else thread.value = null
  }
)

// When the thread refreshes upstream (new reply posted via panel
// inside the drawer), re-pull so the drawer stays in sync. We pass
// down a wrapper that does both: refresh the page list and re-pull
// the thread.
const handleRefresh = () => {
  props.refresh()
  if (props.rootCommentId !== null) loadThread(props.rootCommentId)
}
</script>

<template>
  <KunDrawer
    v-model="isOpen"
    placement="right"
    size="lg"
    title="完整评论线程"
  >
    <KunLoading v-if="isLoading" />

    <KunNull
      v-else-if="!thread"
      description="评论线程不存在或已被删除"
    />

    <GalgameComment
      v-else
      :comment="thread"
      :refresh="handleRefresh"
      :depth="0"
      :max-depth="Infinity"
    />
  </KunDrawer>
</template>
