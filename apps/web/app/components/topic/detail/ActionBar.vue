<script setup lang="ts">
const props = defineProps<{
  topic: TopicDetail
}>()

const { id } = usePersistUserStore()
const { isEdit } = storeToRefs(useTempReplyStore())

// The pill "input" opens the reply editor, which renders as a bottom KunDrawer
// on mobile (see TopicReplyPanel). Floor 0 = replying to the OP.
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
  <!-- Mobile-only floating action pill (desktop keeps the original in-card
       TopicFooter). A frosted, fully-rounded bar that floats a gap above the
       bottom. Visible: input · 收藏 · 评论 · ⋯; the ⋯ holds 推 / 点赞 / 点踩 /
       分享 / 重新编辑 / 隐藏. -->
  <div
    class="border-default/20 bg-background/70 sticky bottom-4 z-30 mx-2 flex items-center gap-1 rounded-full border px-2 py-1.5 shadow-lg backdrop-blur-xl md:hidden"
  >
    <button
      class="bg-default-100 text-default-500 hover:bg-default-200 min-w-0 flex-1 truncate rounded-full px-4 py-2 text-left text-sm transition-colors"
      @click="handleOpenReply"
    >
      发布一条可爱的回复吧～
    </button>

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
        <TopicFooterUpvote
          menu
          :topic-id="topic.id"
          :target-user-id="topic.user.id"
          :upvote-count="topic.upvoteCount"
          :is-upvoted="topic.isUpvoted"
        />
        <TopicFooterLike
          menu
          :topic-id="topic.id"
          :target-user-id="topic.user.id"
          :like-count="topic.likeCount"
          :is-liked="topic.isLiked"
        />
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
