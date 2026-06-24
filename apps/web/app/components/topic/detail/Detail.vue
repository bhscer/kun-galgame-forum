<script setup lang="ts">
import { useTopicReplies } from '~/composables/topic/useTopicReplies'
import { TOPIC_TOC_SOURCE } from '~/composables/topic/useTopicTOC'

const props = defineProps<{
  topic: TopicDetail
}>()

const { id, role } = usePersistUserStore()
const tempReplyStore = useTempReplyStore()
const { lastSuccessfulReply } = storeToRefs(tempReplyStore)
const isTopicAdmin = computed(() => role > 1 || props.topic.user.id === id)

const {
  replies,
  status,
  isComplete,
  sortOrder,
  loadInitialReplies,
  loadMore,
  setSort,
  addNewReply,
  updateReply,
  removeReply
} = useTopicReplies(props.topic.id)

await loadInitialReplies()

provide('topicUserId', props.topic.user.id)

// Feed the TOC rail from data so it renders server-side (no flash on refresh);
// the scrollspy inside useTopicTOC stays client-only.
provide(TOPIC_TOC_SOURCE, {
  getContentHtml: () => props.topic.contentHtml,
  getReplies: () => replies.value
})

watch(
  lastSuccessfulReply,
  (event) => {
    if (!event) {
      return
    }

    switch (event.type) {
      case 'created':
        addNewReply(event.data)

        nextTick(() => {
          document.getElementById(`k${event.data.floor}`)?.scrollIntoView({
            behavior: 'smooth',
            block: 'center'
          })
        })
        break
      case 'updated':
        updateReply(event.data)

        nextTick(() => {
          document.getElementById(`k${event.data.floor}`)?.scrollIntoView({
            behavior: 'smooth',
            block: 'center'
          })
        })
        break
      case 'deleted':
        removeReply(event.data.id)
        break
    }

    tempReplyStore.clearSuccessfulReply()
  },
  { deep: true }
)
</script>

<template>
  <div class="flex flex-col gap-4 lg:flex-row lg:items-start">
    <!-- LEFT: author rail — sticky across the whole page (no card chrome). -->
    <TopicDetailMasterUser v-if="topic.user" :user="topic.user" />

    <!-- RIGHT: post body + 话题小程序 (poll) + tool + replies + action bar. -->
    <div class="min-w-0 flex-1 space-y-4">
      <TopicDetailMaster :topic="topic" />

      <TopicPollContainer :topic-id="topic.id" :is-topic-admin="isTopicAdmin" />

      <!-- scroll-mt offsets the fixed top bar so the 评论 jump button (bottom bar)
           lands the reply-count header at the top, not under the topbar. -->
      <div id="comments-anchor" class="scroll-mt-20">
        <TopicDetailTool
          :reply-count="topic.replyCount"
          :status="status"
          :sort-order="sortOrder"
          @set-sort-order="setSort"
        />
      </div>

      <section id="reply-section" class="space-y-4">
        <div
          v-if="status === 'pending' && replies.length === 0"
          class="flex justify-center py-16"
        >
          <KunLoading description="少女祈祷中..." />
        </div>

        <TopicReplyList
          v-else-if="replies.length > 0"
          :initial-replies="replies"
          :topic-id="topic.id"
          :title="topic.title"
        />

        <div class="py-6 text-center">
          <KunButton
            v-if="!isComplete && status !== 'pending'"
            size="lg"
            variant="flat"
            @click="loadMore"
          >
            加载更多
          </KunButton>
          <KunLoading v-if="status === 'pending' && replies.length > 0" />
          <p v-if="isComplete" class="text-default-500">
            {{ `(｡>︿<｡) 已经一滴回复都不剩了哦~` }}
          </p>
        </div>
      </section>

      <TopicDetailActionBar :topic="topic" />
    </div>
  </div>
</template>
