<script setup lang="ts">
// Rich feed card for TOPIC_CREATION — title · excerpt · first-3 covers · badges
// (NSFW / 有解答 / 投票 / 被推) + section chips · 高赞回复 · a footer row: 收藏 +
// reactions (clickable) on the left, 浏览数 + 查看更多 on the right. Header
// (avatar · username · time) comes from the shared shell.
import { KUN_TOPIC_SECTION } from '~/constants/topic'

const props = defineProps<{ activity: ActivityItem }>()

const data = computed(() => props.activity.data as TopicActivityData | undefined)
const covers = computed(() => (data.value?.coverImages ?? []).slice(0, 3))
const topicId = computed(() => data.value?.topicId ?? 0)
const hasBadge = computed(() => {
  const d = data.value
  return !!d && (d.hasBestAnswer || d.isPoll || d.isNSFW || !!d.upvoteTime)
})

// Per-viewer 收藏 + reaction state (the shared feed can't carry it) — hydrated
// once per session, client-side.
const { isFavorited, reactionKeysOf, ensureLoaded } = useMyTopicInteractions()
onMounted(ensureLoaded)

// Feed reaction counts + the viewer's own "mine" (reactive — patched once the
// hydration lands). Provided to the reaction bar + trigger via reactionsKey; the
// `sync` re-seeds them when "mine" arrives.
const reactionList = computed<KunReaction[]>(() =>
  (data.value?.reactions ?? []).map((r) => ({
    reaction: r.reaction,
    count: r.count,
    mine: reactionKeysOf(topicId.value).includes(r.reaction)
  }))
)
provide(
  reactionsKey,
  useReactions({
    topicId: topicId.value,
    targetUserId: props.activity.actor?.id ?? 0,
    reactions: reactionList.value,
    sync: () => reactionList.value
  })
)
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <div class="space-y-3">
      <KunLink
        underline="none"
        color="default"
        :to="activity.link"
        class-name="group block"
      >
        <div :class="covers.length ? 'flex gap-3' : ''">
          <!-- Cover on the LEFT, vertically centered, no background box. -->
          <div
            v-if="covers.length"
            class="flex w-28 shrink-0 items-center sm:w-40"
          >
            <img
              :src="imageTokenUrl(covers[0]!)"
              alt="话题封面"
              loading="lazy"
              class="max-h-36 w-full rounded-lg object-contain"
            />
          </div>
          <div class="min-w-0 flex-1 space-y-2">
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
          </div>
        </div>
      </KunLink>

      <!-- Badges (NSFW / 有解答 / 投票 / 被推), then the section chips after them. -->
      <div
        v-if="hasBadge || data?.sections?.length"
        class="flex flex-wrap items-center gap-1.5"
      >
        <TopicTagGroup
          v-if="hasBadge"
          :section="[]"
          :tags="[]"
          :upvote-time="data?.upvoteTime"
          :has-best-answer="data?.hasBestAnswer"
          :is-poll-topic="data?.isPoll"
          :is-n-s-f-w-topic="data?.isNSFW"
        />
        <KunChip
          v-for="(sec, index) in data?.sections ?? []"
          :key="index"
          size="sm"
          variant="flat"
          color="primary"
        >
          {{ KUN_TOPIC_SECTION[sec] }}
        </KunChip>
      </div>

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

      <!-- Footer: 收藏 + reactions (clickable) · 浏览 + 查看更多. -->
      <div class="flex items-center justify-between gap-2">
        <div class="flex min-w-0 flex-wrap items-center gap-1">
          <TopicFooterFavorite
            :topic-id="topicId"
            :favorite-count="data?.favoriteCount ?? 0"
            :is-favorite="isFavorited(topicId)"
          />
          <TopicReactionBar />
          <TopicReactionTrigger />
        </div>

        <div
          class="text-default-500 flex shrink-0 items-center gap-3 text-sm"
        >
          <span class="flex items-center gap-1">
            <KunIcon name="lucide:eye" class="size-4" />
            {{ formatNumber(data?.view ?? 0) }}
          </span>
          <KunLink
            underline="none"
            color="default"
            :to="activity.link"
            class-name="text-default-500 hover:text-primary flex items-center gap-0.5 text-sm"
          >
            查看详情
            <KunIcon name="lucide:chevron-right" class="size-4" />
          </KunLink>
        </div>
      </div>
    </div>
  </ActivityCardShell>
</template>
