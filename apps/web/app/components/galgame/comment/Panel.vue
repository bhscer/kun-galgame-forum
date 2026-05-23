<script setup lang="ts">
// Compose-box for a new comment.
//
// Two modes:
//   - Root mode: no parentCommentId. Container renders this above the
//     thread list; targetUserId is the contributor selected from the
//     "评论给" dropdown (commentToUserId in the temp store).
//   - Reply mode: parentCommentId set. Comment.vue renders this inline
//     beneath the comment being replied to; targetUserId is that
//     comment's author.
//
// The two modes share one component — only the wire payload diverges.
const props = defineProps<{
  refresh: () => void
  parentCommentId?: number | null
  targetUserId?: number
}>()

const emits = defineEmits<{
  close: []
}>()

const { commentToUserId } = storeToRefs(useTempGalgameCommentStore())
const route = useRoute()

const content = ref('')
const galgameId = computed(() =>
  parseInt((route.params as { gid: string }).gid)
)
const isPublishing = ref(false)

// Reply mode passes targetUserId explicitly; root mode falls back to
// the dropdown-bound store value.
const effectiveTargetUserId = computed(
  () => props.targetUserId ?? commentToUserId.value
)

const handlePublishComment = async () => {
  if (!content.value.trim()) {
    useMessage(10540, 'warn')
    return
  }
  if (content.value.trim().length > 1007) {
    useMessage(10541, 'warn')
    return
  }

  isPublishing.value = true
  const result = await kunFetch(`/galgame/${galgameId.value}/comment`, {
    method: 'POST',
    body: {
      galgameId: galgameId.value,
      targetUserId: effectiveTargetUserId.value,
      parentCommentId: props.parentCommentId ?? null,
      content: content.value
    }
  })
  isPublishing.value = false

  if (result) {
    content.value = ''
    useMessage(10542, 'success')
    emits('close')
    props.refresh()
  }
}
</script>

<template>
  <div class="space-y-3">
    <KunTextarea
      v-model="content"
      :placeholder="
        parentCommentId
          ? '写下您的回复'
          : '请注意您 “评论给” 的用户, 只有被评论的用户才会收到您的评论通知, 因此您需要在 “评论给” 的用户中选择一位资源发布者或贡献者'
      "
      name="comment"
      :rows="parentCommentId ? 3 : 5"
    />

    <div class="flex items-center justify-between">
      <slot />

      <KunButton
        class="ml-auto"
        :loading="isPublishing"
        @click="handlePublishComment"
      >
        {{ parentCommentId ? '发布回复' : '发布评论' }}
      </KunButton>
    </div>
  </div>
</template>
