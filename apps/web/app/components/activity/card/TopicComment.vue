<script setup lang="ts">
// TOPIC_COMMENT_CREATION — a comment on a reply. Same shape as the reply card:
// the comment body (a few lines), the reply it's on (被评论的评论) quoted above,
// and the topic name anchored at the bottom.
const props = defineProps<{ activity: ActivityItem }>()

const data = computed(
  () => props.activity.data as TopicCommentActivityData | undefined
)
const quoted = computed(() => data.value?.quotedReply)
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <KunLink
      underline="none"
      color="default"
      :to="activity.link"
      class-name="group block space-y-2"
    >
      <!-- The reply being commented on (被评论的评论). -->
      <ActivityCardQuote
        v-if="quoted"
        :content="quoted.content"
        :label="`#${quoted.floor}`"
      />

      <p
        class="group-hover:text-primary line-clamp-4 text-base break-all transition-colors"
      >
        {{ markdownToText(activity.content) }}
      </p>

      <p
        v-if="data?.topicTitle"
        class="text-default-500 flex items-center gap-1 text-sm"
      >
        <KunIcon name="icon-park-outline:topic" class="size-4 shrink-0" />
        <span class="line-clamp-1">{{ data.topicTitle }}</span>
      </p>
    </KunLink>
  </ActivityCardShell>
</template>
