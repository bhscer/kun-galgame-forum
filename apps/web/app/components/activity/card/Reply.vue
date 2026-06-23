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
    <div class="space-y-2">
      <!-- The reply being replied to (quoted #floor). -->
      <ActivityCardQuote
        v-if="quoted"
        :content="quoted.content"
        :label="`#${quoted.floor}`"
      />

      <!-- Body sits OUTSIDE the link so its 显示更多 toggle never navigates;
           preserveNewlines + pre-line keep the author's line breaks. -->
      <ActivityCollapse :max-height="300">
        <p class="text-default-700 text-base break-all whitespace-pre-line">
          {{ markdownToText(activity.content, { preserveNewlines: true }) }}
        </p>
      </ActivityCollapse>

      <KunLink
        v-if="data?.topicTitle"
        underline="none"
        color="default"
        :to="activity.link"
        class-name="text-default-500 hover:text-primary flex items-center gap-1 text-sm"
      >
        <KunIcon name="icon-park-outline:topic" class="size-4 shrink-0" />
        <span class="line-clamp-1">{{ data.topicTitle }}</span>
      </KunLink>
    </div>
  </ActivityCardShell>
</template>
