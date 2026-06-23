<script setup lang="ts">
import { useTopic } from '~/composables/topic/useTopic'

const { topics, isLoadingComplete, isFetching, loadInitialTopics } =
  useTopic('all')

await loadInitialTopics()
</script>

<template>
  <div class="space-y-3">
    <KunHeader
      name="话题列表"
      description="鲲 Galgame 论坛的全部话题，涵盖 Galgame 讨论、技术交流、资源求助与日常闲聊，在这里和大家一起畅所欲言。"
    />

    <!-- List layout: each topic separated by a faint divider, no card chrome. -->
    <div class="divide-default-200/60 divide-y">
      <TopicCard v-for="topic in topics" :key="topic.id" :topic="topic" />
    </div>

    <div class="flex w-full items-center justify-center p-6">
      <KunLoading v-if="isFetching" description="正在摸鱼中...咕咕咕" />
      <KunNull v-if="isLoadingComplete" description="真的一滴也不剩了呜呜呜" />
    </div>
  </div>
</template>
