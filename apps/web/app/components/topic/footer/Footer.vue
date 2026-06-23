<script setup lang="ts">
defineProps<{
  topic: TopicDetail
}>()

const { id } = usePersistUserStore()
</script>

<template>
  <!-- Desktop only: on mobile the floating TopicDetailActionBar replaces this. -->
  <div class="mt-auto hidden items-center justify-between md:flex">
    <div class="flex items-center gap-1">
      <TopicFooterUpvote
        :topic-id="topic.id"
        :target-user-id="topic.user.id"
        :upvote-count="topic.upvoteCount"
        :is-upvoted="topic.isUpvoted"
      />

      <TopicFooterFavorite
        :topic-id="topic.id"
        :target-user-id="topic.user.id"
        :favorite-count="topic.favoriteCount"
        :is-favorite="topic.isFavorited"
      />

      <TopicReactionTrigger />
    </div>

    <div class="flex items-center gap-1">
      <TopicFooterReply
        :target-user-name="topic.user.name"
        :target-user-id="topic.user.id"
        :target-floor="0"
      />

      <KunTooltip text="分享">
        <KunButton
          :is-icon-only="true"
          variant="light"
          color="default"
          size="lg"
          @click="
            useKunCopy(
              `${topic.title}: https://www.kungal.com/topic/${topic.id}`
            )
          "
        >
          <KunIcon name="lucide:share-2" />
        </KunButton>
      </KunTooltip>

      <TopicFooterRewrite :topic="topic" />

      <KunPopover position="top-end">
        <template v-if="id" #trigger>
          <KunButton
            :is-icon-only="true"
            variant="light"
            color="default"
            size="lg"
          >
            <KunIcon name="lucide:ellipsis" />
          </KunButton>
        </template>

        <div class="flex w-54 flex-col gap-2 p-2">
          <TopicFooterHide :topic-id="topic.id" />
        </div>
      </KunPopover>
    </div>
  </div>
</template>
