<script setup lang="ts">
// Rich feed card for TOPIC_REPLY_CREATION:
//   top    — username + time (shell)
//   middle — the reply body (a few lines); if it quoted another reply, that
//            quoted reply shows as a nested block above the body
//   bottom — the title of the topic the reply is in
// content + the quoted body arrive with @/# tokens already resolved (BE).
const props = defineProps<{ activity: ActivityItem }>()

const data = computed(() => props.activity.data as ReplyActivityData | undefined)
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
      <!-- The reply being replied to (quoted #floor). -->
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
