<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

// Recursive node in the threaded comment tree.
//
// Visible recursion stops at MAX_DEPTH. At that depth, any further
// children are not rendered inline — instead a "查看更多 N 条回复"
// button bubbles an open-thread event up to the container, which opens
// the ThreadDrawer with the entire root thread. This matches the spec:
// readable depth inline, full exploration in a focused drawer.
const props = withDefaults(
  defineProps<{
    comment: SerializeObject<GalgameComment>
    refresh: () => void
    depth?: number
    // Visible recursion cap. Inline list passes 2 (renders depth
    // 0/1/2 = three layers, anything deeper hits the load-more
    // button). ThreadDrawer passes Infinity so the focused view is
    // depth-unbounded.
    maxDepth?: number
  }>(),
  { depth: 0, maxDepth: 2 }
)

const emit = defineEmits<{
  openThread: [rootId: number]
}>()

const galgame = inject<GalgameDetail>('galgame')
const { id, role } = usePersistUserStore()

const isShowReply = ref(false)
const isShowDelete = computed(
  () => props.comment.user?.id === id || galgame?.user.id === id || role >= 2
)

const visibleReplies = computed(() => {
  if (props.depth >= props.maxDepth) return []
  return props.comment.replies ?? []
})

// At max depth, surface a button if any descendants exist; the count
// shown is the comment's transitive descendant count, computed server-
// side so it stays accurate across reloads.
const showLoadMore = computed(
  () =>
    props.depth >= props.maxDepth &&
    (props.comment.replies?.length ?? 0) > 0 &&
    props.comment.replyCount > 0
)

const handleDelete = async () => {
  const ok = await useComponentMessageStore().alert('您确定删除评论吗？')
  if (!ok) return

  const result = await kunFetch(`/galgame/${props.comment.galgameId}/comment`, {
    method: 'DELETE',
    query: { commentId: props.comment.id }
  })

  if (result) {
    useMessage(10538, 'success')
    props.refresh()
  }
}

const handleOpenThread = () => {
  // The root anchor: a reply's rootCommentId points to its thread root;
  // a root comment (rootCommentId === null) IS the root.
  emit('openThread', props.comment.rootCommentId ?? props.comment.id)
}
</script>

<template>
  <div class="flex gap-3">
    <KunAvatar :user="comment.user" size="sm" />

    <div class="min-w-0 flex-1 space-y-1.5">
      <div class="flex flex-wrap items-baseline gap-1.5">
        <span class="text-default-800 text-sm font-medium">
          {{ comment.user.name }}
        </span>
        <template v-if="comment.targetUser">
          <KunIcon
            name="lucide:arrow-right"
            class="text-default-400 text-xs"
          />
          <KunLink
            underline="hover"
            size="sm"
            :to="`/user/${comment.targetUser.id}/info`"
          >
            {{ comment.targetUser.name }}
          </KunLink>
        </template>
        <span class="text-default-400 text-xs">
          {{ formatTimeDifference(comment.created) }}
        </span>
      </div>

      <p class="text-default-700 text-sm break-all whitespace-pre-line">
        {{ comment.content }}
      </p>

      <div class="-ml-2 flex items-center gap-1">
        <KunButton
          variant="light"
          size="sm"
          class-name="gap-1"
          @click="isShowReply = !isShowReply"
        >
          <KunIcon name="lucide:reply" />
          回复
        </KunButton>

        <GalgameCommentLike :comment="comment" />

        <KunTooltip v-if="isShowDelete" text="删除">
          <KunButton
            :is-icon-only="true"
            variant="light"
            color="danger"
            size="sm"
            @click="handleDelete"
          >
            <KunIcon name="lucide:trash-2" />
          </KunButton>
        </KunTooltip>
      </div>

      <KunAnimationFadeCard>
        <GalgameCommentPanel
          v-if="isShowReply"
          :parent-comment-id="comment.id"
          :target-user-id="comment.user.id"
          :refresh="refresh"
          @close="isShowReply = false"
        />
      </KunAnimationFadeCard>

      <div
        v-if="visibleReplies.length"
        class="border-default-200 mt-3 space-y-4 border-l pl-4"
      >
        <GalgameComment
          v-for="reply in visibleReplies"
          :key="reply.id"
          :comment="reply"
          :refresh="refresh"
          :depth="depth + 1"
          :max-depth="maxDepth"
          @open-thread="(rootId) => emit('openThread', rootId)"
        />
      </div>

      <KunButton
        v-if="showLoadMore"
        variant="light"
        color="primary"
        size="sm"
        full-width
        class-name="mt-2"
        @click="handleOpenThread"
      >
        <KunIcon name="lucide:messages-square" />
        查看更多 {{ comment.replyCount }} 条回复
      </KunButton>
    </div>
  </div>
</template>
