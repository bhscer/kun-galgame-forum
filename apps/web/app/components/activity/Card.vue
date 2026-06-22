<script setup lang="ts">
// One feed item — dispatches to the rich card for its activity type, falling
// back to the generic card for types without one yet (forward-compatible: a new
// type renders fine until it gets a rich card). Add a branch here + a card under
// ./card/ to enrich another type. All variants share the feed's spacing/chrome.
defineProps<{ activity: ActivityItem }>()

// Galgame-scoped types that share the compact "references a galgame" card.
// (GALGAME_EDIT and GALGAME_RATING_CREATION have their own richer cards.)
const GALGAME_REF_TYPES = new Set([
  'GALGAME_PR_CREATION',
  'GALGAME_COMMENT_CREATION',
  'GALGAME_RESOURCE_CREATION'
])
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
  <ActivityCardGalgameRef
    v-else-if="GALGAME_REF_TYPES.has(activity.type) && activity.data"
    :activity="activity"
  />
  <ActivityCardGeneric v-else :activity="activity" />
</template>
