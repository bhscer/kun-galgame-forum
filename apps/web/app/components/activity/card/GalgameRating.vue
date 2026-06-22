<script setup lang="ts">
// GALGAME_RATING_CREATION — rating feed card. Layout like the new-galgame card:
// cover left; right = game title, then <通关状态> <推荐程度> <总评分>, then the
// 简评 (or a spoiler notice), then a 点赞 button + 查看详情. Labels mirror the
// galgame rating detail row.
import {
  KUN_GALGAME_RATING_PLAY_STATUS_MAP,
  KUN_GALGAME_RATING_RECOMMEND_MAP,
  KUN_GALGAME_RATING_RECOMMEND_COLOR_MAP,
  KUN_GALGAME_RATING_SPOILER_WARNING
} from '~/constants/galgame-rating'

const props = defineProps<{ activity: ActivityItem }>()

const data = computed(
  () => props.activity.data as GalgameActivityData | undefined
)
const rating = computed(() => data.value?.rating)
const gid = computed(() => data.value?.galgameId ?? 0)
const galgameLink = computed(() =>
  gid.value ? `/galgame/${gid.value}` : props.activity.link
)

const playStatusLabel = computed(() =>
  rating.value
    ? KUN_GALGAME_RATING_PLAY_STATUS_MAP[rating.value.playStatus] ||
      rating.value.playStatus
    : ''
)
const recommendLabel = computed(() =>
  rating.value
    ? KUN_GALGAME_RATING_RECOMMEND_MAP[rating.value.recommend] ||
      rating.value.recommend
    : ''
)
const recommendColor = computed(() => {
  const c = rating.value
    ? KUN_GALGAME_RATING_RECOMMEND_COLOR_MAP[rating.value.recommend]
    : ''
  switch (c) {
    case 'danger':
      return 'text-danger'
    case 'success':
      return 'text-success'
    case 'warning':
      return 'text-warning'
    case 'secondary':
      return 'text-secondary'
    default:
      return 'text-default-600'
  }
})
const overall = computed(() => rating.value?.overall.toFixed(1) ?? '')
const hasSpoiler = computed(
  () => !!rating.value && rating.value.spoilerLevel !== 'none'
)
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <div class="space-y-3">
      <div class="flex items-start gap-3">
        <KunLink :to="galgameLink" class-name="shrink-0">
          <div
            class="bg-default-100 aspect-video w-32 overflow-hidden rounded-lg sm:w-44"
          >
            <img
              v-if="data?.coverHash"
              :src="imageHashUrl(imageCdnBase(), data.coverHash, 'mini')"
              :alt="data?.name"
              loading="lazy"
              class="h-full w-full object-cover"
            />
          </div>
        </KunLink>

        <div class="min-w-0 flex-1 space-y-1.5">
          <KunLink
            underline="none"
            color="default"
            :to="galgameLink"
            class-name="hover:text-primary block"
          >
            <h3 class="line-clamp-2 font-medium break-all">{{ data?.name }}</h3>
          </KunLink>

          <div
            v-if="rating"
            class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm"
          >
            <span class="text-default-700">{{ playStatusLabel }}</span>
            <span :class="cn('font-medium', recommendColor)">
              {{ recommendLabel }}
            </span>
            <span
              class="text-default-800 inline-flex items-center gap-1 font-semibold"
            >
              <KunIcon name="lucide:star" class="text-warning size-4" />
              {{ overall }}
            </span>
          </div>

          <p
            v-if="hasSpoiler"
            class="text-default-500 inline-flex items-center gap-1 text-sm"
          >
            <KunIcon name="lucide:eye-off" class="size-4 shrink-0" />
            {{ KUN_GALGAME_RATING_SPOILER_WARNING }}
          </p>
          <p
            v-else-if="rating?.shortSummary"
            class="text-default-500 line-clamp-3 text-sm break-all"
          >
            {{ rating.shortSummary }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <GalgameRatingDetailLike
          :rating-id="rating?.ratingId"
          :target-user-id="rating?.authorId ?? activity.actor?.id ?? 0"
          :like-count="rating?.likeCount ?? 0"
          :is-liked="false"
        />
        <KunLink
          underline="none"
          color="default"
          :to="activity.link"
          class-name="text-default-500 hover:text-primary ml-auto flex items-center gap-0.5 text-sm"
        >
          查看详情
          <KunIcon name="lucide:chevron-right" class="size-4" />
        </KunLink>
      </div>
    </div>
  </ActivityCardShell>
</template>
