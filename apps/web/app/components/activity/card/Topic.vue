<script setup lang="ts">
// Rich feed card for TOPIC_CREATION — title · excerpt · first-3 covers · badges
// (该话题被推 / 有解答 / 投票 / NSFW via TopicTagGroup) · 高赞回复 · stat row (分区 +
// 浏览/赞/回复). Header (avatar · username · time) comes from the shared shell.
import { KUN_TOPIC_SECTION } from '~/constants/topic'

const props = defineProps<{ activity: ActivityItem }>()

const data = computed(() => props.activity.data as TopicActivityData | undefined)
const covers = computed(() => (data.value?.coverImages ?? []).slice(0, 3))
const hasBadge = computed(() => {
  const d = data.value
  return !!d && (d.hasBestAnswer || d.isPoll || d.isNSFW || !!d.upvoteTime)
})
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <div class="space-y-3">
      <KunLink
        underline="none"
        color="default"
        :to="activity.link"
        class-name="group block space-y-2.5"
      >
        <h3
          class="group-hover:text-primary line-clamp-2 text-base font-medium break-all transition-colors"
        >
          {{ activity.content }}
        </h3>

        <p
          v-if="data?.excerpt"
          class="text-default-500 line-clamp-3 text-sm break-all"
        >
          {{ markdownToText(data.excerpt) }}
        </p>

        <TopicCoverGrid v-if="covers.length" :images="covers" />
      </KunLink>

      <TopicTagGroup
        v-if="hasBadge"
        :section="[]"
        :tags="[]"
        :upvote-time="data?.upvoteTime"
        :has-best-answer="data?.hasBestAnswer"
        :is-poll-topic="data?.isPoll"
        :is-n-s-f-w-topic="data?.isNSFW"
      />

      <KunLink
        v-if="data?.topReply"
        underline="none"
        color="default"
        :to="activity.link"
        class-name="border-primary/40 bg-default-100/50 text-default-600 hover:bg-default-100 flex items-start gap-2 rounded-md border-l-2 px-2 py-1.5 text-sm"
      >
        <KunIcon
          name="lucide:message-circle-heart"
          class="mt-0.5 size-4 shrink-0"
        />
        <span class="line-clamp-2 min-w-0 flex-1 break-all">
          {{ markdownToText(data.topReply.content) }}
        </span>
        <span class="text-default-500 flex shrink-0 items-center gap-1">
          <KunIcon name="lucide:thumbs-up" class="size-3.5" />
          {{ data.topReply.likeCount }}
        </span>
      </KunLink>

      <div
        class="text-default-500 flex flex-wrap items-center gap-x-6 gap-y-2 text-sm"
      >
        <div v-if="data?.sections?.length" class="flex flex-wrap gap-1.5">
          <KunChip
            v-for="(sec, index) in data.sections"
            :key="index"
            size="sm"
            variant="flat"
            color="primary"
          >
            {{ KUN_TOPIC_SECTION[sec] }}
          </KunChip>
        </div>

        <div class="flex items-center gap-4">
          <span class="flex items-center gap-1">
            <KunIcon name="lucide:eye" class="h-4 w-4" />
            {{ formatNumber(data?.view ?? 0) }}
          </span>
          <span class="flex items-center gap-1">
            <KunIcon name="lucide:thumbs-up" class="h-4 w-4" />
            {{ data?.likeCount ?? 0 }}
          </span>
          <span class="flex items-center gap-1">
            <KunIcon name="carbon:reply" class="h-4 w-4" />
            {{ (data?.replyCount ?? 0) + (data?.commentCount ?? 0) }}
          </span>
        </div>
      </div>
    </div>
  </ActivityCardShell>
</template>
