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
  <ActivityCardGeneric v-else :activity="activity" />
</template>
