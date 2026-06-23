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
        <KunReaction
          :toggle="false"
          icon="lucide:share-2"
          label="分享"
          @click="
            useKunCopy(
              `${topic.title}: https://www.kungal.com/topic/${topic.id}`
            )
          "
        />
      </KunTooltip>

      <TopicFooterRewrite :topic="topic" />

      <KunPopover position="top-end">
        <template v-if="id" #trigger>
          <KunReaction :toggle="false" icon="lucide:ellipsis" label="更多" />
        </template>

        <div class="flex w-54 flex-col gap-2 p-2">
          <TopicFooterHide :topic-id="topic.id" />
        </div>
      </KunPopover>
    </div>
  </div>
</template>
