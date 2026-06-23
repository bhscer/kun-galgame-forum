<script setup lang="ts">
const props = defineProps<{
  topic: TopicDetail
}>()

// Shared reaction state for this topic: the chips (below) + the trigger (in the
// footer on desktop) both inject this, so reacting in either place stays in sync.
provide(
  reactionsKey,
  useReactions({
    topicId: props.topic.id,
    targetUserId: props.topic.user.id,
    reactions: props.topic.reactions
  })
)
</script>

<template>
  <div
    :class="
      cn(
        'outline-primary flex justify-between gap-3 rounded-lg outline-offset-2'
      )
    "
    id="0"
  >
    <TopicDetailMasterUser v-if="topic.user" :user="topic.user" />

    <KunCard
      :is-transparent="false"
      :is-hoverable="false"
      class-name="lg:w-[calc(100%-220px)] w-full min-w-0"
      content-class="gap-4 justify-start"
    >
      <!-- Post header — the title leads the hierarchy (larger than any in-body
           heading), with categorization chips and a compact icon byline below. -->
      <header class="space-y-3">
        <h1
          class="text-3xl leading-tight font-bold tracking-tight break-words lg:text-4xl"
        >
          {{ topic.title }}
        </h1>

        <TopicTagGroup
          :section="topic.section"
          :tags="topic.tag"
          :upvote-time="topic.upvoteTime"
          :has-best-answer="false"
          :is-poll-topic="topic.isPollTopic"
          :is-n-s-f-w-topic="topic.isNSFW"
          :is-nav-to-section="true"
        />

        <div
          class="text-default-500 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm"
        >
          <span class="flex items-center gap-1.5">
            <KunIcon name="lucide:eye" class="size-4" />
            {{ topic.view }}
          </span>
          <span class="flex items-center gap-1.5">
            <KunIcon name="lucide:clock" class="size-4" />
            <KunTime :time="topic.created" type="datetime" show-year />
          </span>
          <span v-if="topic.edited" class="flex items-center gap-1.5">
            <KunIcon name="lucide:pencil-line" class="size-4" />
            编辑于 <KunTime :time="topic.edited" type="datetime" show-year />
          </span>
        </div>
      </header>

      <TopicDetailBestAnswer
        v-if="topic.bestAnswer"
        :best-answer="topic.bestAnswer"
      />

      <TopicDetailUser
        class-name="lg:hidden"
        :user="topic.user"
        :created="topic.created"
        :edited="topic.edited"
        :topic-id="topic.id"
        :floor="0"
        :show-addition="false"
      />

      <KunDivider />

      <KunContent
        class="kun-master"
        :content="renderKatex(topic.contentHtml)"
      />

      <KunDivider />

      <div class="flex flex-wrap items-center gap-1.5">
        <TopicReactionBar />
        <!-- Desktop shows the trigger in the footer (next to favorite). -->
        <TopicReactionTrigger class="md:hidden" />
      </div>

      <p class="text-default-500 ml-auto text-sm">
        本文版权遵循
        <KunLink
          underline="hover"
          size="sm"
          class-name="text-default-500"
          target="_blank"
          rel="noopener noreferrer"
          to="https://creativecommons.org/licenses/by-nc/4.0/deed.en"
        >
          CC BY-NC 协议
        </KunLink>
        和
        <KunLink
          underline="hover"
          size="sm"
          class-name="text-default-500"
          to="/doc/article-copyright"
        >
          本站版权政策
        </KunLink>
      </p>

      <TopicFooter :topic="topic" />
    </KunCard>
  </div>
</template>
