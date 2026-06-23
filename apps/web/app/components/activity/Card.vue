<script setup lang="ts">
// One feed item — dispatches to the rich card for its activity type, falling
// back to the generic card for types without one yet (forward-compatible: a new
// type renders fine until it gets a rich card). Add a branch here + a card under
// ./card/ to enrich another type. All variants share the feed's spacing/chrome.
defineProps<{ activity: ActivityItem }>()
</script>

<template>
  <ActivityCardTopic
    v-if="activity.type === 'TOPIC_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardReply
    v-else-if="activity.type === 'TOPIC_REPLY_CREATION'"
    :activity="activity"
  />
  <ActivityCardGalgame
    v-else-if="activity.type === 'GALGAME_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardGalgameEdit
    v-else-if="activity.type === 'GALGAME_EDIT' && activity.data"
    :activity="activity"
  />
  <ActivityCardGalgameRating
    v-else-if="activity.type === 'GALGAME_RATING_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardGalgameComment
    v-else-if="activity.type === 'GALGAME_COMMENT_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardGalgameResource
    v-else-if="activity.type === 'GALGAME_RESOURCE_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardTopicComment
    v-else-if="activity.type === 'TOPIC_COMMENT_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardGalgamePr
    v-else-if="activity.type === 'GALGAME_PR_CREATION' && activity.data"
    :activity="activity"
  />
  <ActivityCardGeneric v-else :activity="activity" />
</template>
