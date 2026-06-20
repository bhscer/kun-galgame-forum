<script setup lang="ts">
const props = defineProps<{
  topic: TopicDetail
}>()

const { id } = usePersistUserStore()
const { isEdit } = storeToRefs(useTempReplyStore())

// Both the mobile pill "input" and the desktop 回复 button open the reply
// editor. On mobile that editor renders as a bottom KunDrawer, on desktop as
// the fixed panel (see TopicReplyPanel). Floor 0 = replying to the OP.
const handleOpenReply = () => {
  if (!id) {
    useAuthModal().open()
    return
  }
  isEdit.value = true
}

const scrollToComments = () => {
  document
    .getElementById('comments-anchor')
    ?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const handleShare = () => {
  useKunCopy(
    `${props.topic.title}: https://www.kungal.com/topic/${props.topic.id}`
  )
}
</script>

<template>
  <!-- Sticky bottom action bar (app-style). One flex row, responsive: mobile
       leads with a pill input (→ reply drawer); desktop drops the input and
       adds a labeled 回复 button on the right. Visible: 推 / 点赞 / 收藏 / 评论;
       ⋯ holds 点踩 / 分享 / 重新编辑 / 隐藏. -->
  <div
    class="border-default/20 bg-background/90 sticky bottom-0 z-30 flex items-center gap-1 rounded-t-2xl border-t px-3 py-2 shadow-[0_-1px_12px_rgba(0,0,0,0.05)] backdrop-blur-[var(--kun-background-blur)]"
  >
    <button
      class="bg-default-100 text-default-500 hover:bg-default-200 min-w-0 flex-1 truncate rounded-full px-4 py-2 text-left text-sm transition-colors md:hidden"
      @click="handleOpenReply"
    >
      发布一条可爱的回复吧～
    </button>

    <TopicFooterUpvote
      :topic-id="topic.id"
      :target-user-id="topic.user.id"
      :upvote-count="topic.upvoteCount"
      :is-upvoted="topic.isUpvoted"
    />
    <TopicFooterLike
      :topic-id="topic.id"
      :target-user-id="topic.user.id"
      :like-count="topic.likeCount"
      :is-liked="topic.isLiked"
    />
    <TopicFooterFavorite
      :topic-id="topic.id"
      :favorite-count="topic.favoriteCount"
      :is-favorite="topic.isFavorited"
    />

    <KunTooltip text="跳到评论区">
      <KunButton
        :is-icon-only="true"
        variant="light"
        color="default"
        size="lg"
        @click="scrollToComments"
      >
        <KunIcon name="uil:comment-dots" />
      </KunButton>
    </KunTooltip>

    <!-- Desktop spacer pushes 回复 + ⋯ to the far right; collapses on mobile. -->
    <div class="hidden md:block md:flex-1" />

    <KunButton
      class-name="hidden md:flex"
      variant="flat"
      @click="handleOpenReply"
    >
      回复
    </KunButton>

    <KunPopover position="top-end" inner-class="p-2 w-56">
      <template #trigger>
        <KunButton
          :is-icon-only="true"
          variant="light"
          color="default"
          size="lg"
        >
          <KunIcon name="lucide:ellipsis" />
        </KunButton>
      </template>

      <div class="flex flex-col gap-1">
        <TopicFooterDislike
          menu
          :topic-id="topic.id"
          :target-user-id="topic.user.id"
          :dislike-count="topic.dislikeCount"
          :is-disliked="topic.isDisliked"
        />
        <KunButton
          variant="light"
          color="default"
          size="sm"
          class-name="w-full justify-start gap-2 whitespace-nowrap"
          @click="handleShare"
        >
          <KunIcon class-name="text-lg" name="lucide:share-2" />
          分享
        </KunButton>
        <TopicFooterRewrite menu :topic="topic" />
        <TopicFooterHide :topic-id="topic.id" />
      </div>
    </KunPopover>
  </div>
</template>
