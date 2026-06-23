<script setup lang="ts">
const props = defineProps<{
  topic: TopicCard
}>()

const actionsCount = computed(
  () => props.topic.replyCount + props.topic.commentCount
)
</script>

<template>
  <NuxtLink
    :to="`/topic/${topic.id}`"
    class="group block space-y-2 py-4 first:pt-0 last:pb-0"
  >
    <h3
      class="group-hover:text-primary line-clamp-2 text-lg font-medium transition-colors"
    >
      {{ topic.title }}
    </h3>

    <TopicTagGroup
      :section="props.topic.section"
      :tags="props.topic.tag"
      :has-best-answer="topic.hasBestAnswer"
      :is-poll-topic="topic.isPollTopic"
      :is-n-s-f-w-topic="topic.isNSFWTopic"
    />

    <div class="text-default-700 flex items-center gap-4 text-sm">
      <span class="flex items-center gap-1">
        <KunIcon class="size-4" name="lucide:eye" />
        {{ formatNumber(props.topic.view) }}
      </span>

      <span v-if="props.topic.likeCount" class="flex items-center gap-1">
        <KunIcon class="size-4" name="lucide:thumbs-up" />
        {{ props.topic.likeCount }}
      </span>

      <span v-if="actionsCount" class="flex items-center gap-1">
        <KunIcon class="size-4" name="carbon:reply" />
        {{ actionsCount }}
      </span>
    </div>

    <!-- Footer (left → right): avatar · name · publish time (relative within a
         day, otherwise a precise date). -->
    <div class="text-default-600 flex items-center gap-2 text-sm">
      <KunAvatar
        :disable-floating="true"
        :user="topic.user"
        size="xs"
        :is-navigation="false"
      />
      <span>{{ topic.user.name }}</span>
      <KunTime :time="topic.created" type="auto" />
    </div>
  </NuxtLink>
</template>
