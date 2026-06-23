<script setup lang="ts">
defineProps<{
  title: string
  reply: TopicReply
}>()

const emits = defineEmits<{
  handleNewComment: [comment: TopicComment]
}>()

const { id } = usePersistUserStore()
const isCommentPanelVisible = ref(false)

const handleClickComment = () => {
  if (!id) {
    useAuthModal().open()
    return
  }
  isCommentPanelVisible.value = !isCommentPanelVisible.value
}

const handleNewComment = (comment: TopicComment) => {
  emits('handleNewComment', comment)
  isCommentPanelVisible.value = false
}
</script>

<template>
  <div class="w-full">
    <div class="flex items-center justify-end">
      <div class="flex items-center gap-1">
        <TopicFooterReply
          :target-user-name="reply.user.name"
          :target-user-id="reply.user.id"
          :target-floor="reply.floor"
          :target-reply-id="reply.id"
        />
        <KunTooltip text="分享该回复">
          <KunReaction
            :toggle="false"
            icon="lucide:share-2"
            label="分享该回复"
            @click="
              useKunCopy(
                `${title}: https://www.kungal.com/topic/${reply.topicId}#k${reply.floor}`
              )
            "
          />
        </KunTooltip>
        <TopicReplyRewrite :reply="reply" />
        <KunTooltip text="评论">
          <KunReaction
            :toggle="false"
            icon="uil:comment-dots"
            label="评论"
            @click="handleClickComment"
          />
        </KunTooltip>
        <!-- ... 更多按钮 ... -->
        <KunPopover position="top-end">
          <template v-if="id" #trigger>
            <KunReaction :toggle="false" icon="lucide:ellipsis" label="更多" />
          </template>

          <div class="flex w-54 flex-col gap-2 p-2">
            <TopicReplyPin :reply="reply" />
            <TopicReplyBestAnswer :reply="reply" />
            <TopicReplyDelete :reply="reply" />
          </div>
        </KunPopover>
      </div>
    </div>

    <KunFadeCard>
      <LazyTopicCommentPanel
        v-if="isCommentPanelVisible"
        class="mt-4"
        :reply-id="reply.id"
        :target-user="reply.user"
        @get-comment="handleNewComment"
        @close-panel="isCommentPanelVisible = false"
      />
    </KunFadeCard>
  </div>
</template>
