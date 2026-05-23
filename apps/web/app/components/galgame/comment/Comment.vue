<script setup lang="ts">
import type { SerializeObject } from 'nitropack'

// One node in the comment view. The visual model is FLAT — two tiers:
//
//   depth 0 = root (top-level comment)
//   depth 1 = reply (anything in the root's flat replies list)
//
// We never render past depth 1. The DB still tracks the true parent
// chain via parent_comment_id, but the inline list and the drawer
// both render replies as one flat group beneath the root. The
// "回复 @<user>" context survives via comment.targetUser ("A => B").
//
// Mutations bubble up as events; Container / ThreadDrawer own the
// state and apply immutable updates.
const props = withDefaults(
  defineProps<{
    comment: SerializeObject<GalgameComment>
    depth?: number
  }>(),
  { depth: 0 }
)

const emit = defineEmits<{
  openThread: [rootId: number]
  replyAdded: [reply: GalgameComment]
  // Carries rootId so the controller can locate the affected root
  // even when the deleted comment isn't in the loaded slice (e.g. a
  // hidden grandchild deleted via the drawer).
  replyRemoved: [commentId: number, removedSubtreeSize: number, rootId: number]
}>()

const galgame = inject<GalgameDetail>('galgame')
const { id, role } = usePersistUserStore()

const isShowReply = ref(false)
const isShowDelete = computed(
  () => props.comment.user?.id === id || galgame?.user.id === id || role >= 2
)

// Only the root has visible children; replies render leaf-only.
const visibleReplies = computed(() =>
  props.depth === 0 ? props.comment.replies ?? [] : []
)

// "查看更多 N 条回复" — only ever appears under a root that has
// more total descendants than what's been loaded into .replies. When
// the drawer is mounted with the full thread, replies.length equals
// replyCount, so the button vanishes naturally.
const showLoadMore = computed(
  () =>
    props.depth === 0 &&
    (props.comment.replyCount ?? 0) > visibleReplies.value.length
)

const rootAnchorId = computed(
  () => props.comment.rootCommentId ?? props.comment.id
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
    const removed = 1 + (props.comment.replyCount ?? 0)
    emit('replyRemoved', props.comment.id, removed, rootAnchorId.value)
  }
}
</script>

<template>
  <div class="flex gap-3">
    <KunAvatar :user="comment.user" :size="depth === 0 ? 'md' : 'sm'" />

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
          @close="isShowReply = false"
          @submitted="(reply) => emit('replyAdded', reply)"
        />
      </KunAnimationFadeCard>

      <!-- Replies render flush — single visual tier, no indent or
           border-l. Smaller avatar on the children is the only
           hierarchy cue. -->
      <div v-if="visibleReplies.length" class="mt-3 space-y-4">
        <GalgameComment
          v-for="reply in visibleReplies"
          :key="reply.id"
          :comment="reply"
          :depth="1"
          @reply-added="(r) => emit('replyAdded', r)"
          @reply-removed="(id, size, rootId) => emit('replyRemoved', id, size, rootId)"
        />
      </div>

      <KunButton
        v-if="showLoadMore"
        variant="light"
        color="primary"
        size="sm"
        full-width
        class-name="mt-2"
        @click="emit('openThread', rootAnchorId)"
      >
        <KunIcon name="lucide:messages-square" />
        查看更多 {{ comment.replyCount }} 条回复
      </KunButton>
    </div>
  </div>
</template>
