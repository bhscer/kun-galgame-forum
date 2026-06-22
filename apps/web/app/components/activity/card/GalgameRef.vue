<script setup lang="ts">
// Compact card for galgame-scoped activity that REFERENCES a galgame — edit / PR
// / comment / rating / resource. Embedded galgame preview (cover left) with the
// action text on the right (activity.content is already "编辑了《X》" … or the raw
// comment); the cover carries the galgame's visual identity (no type chip).
const props = defineProps<{ activity: ActivityItem }>()

const data = computed(
  () => props.activity.data as GalgameActivityData | undefined
)
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <ActivityCardGalgamePreview
      :to="activity.link"
      :cover-hash="data?.coverHash"
      :name="data?.name"
    >
      <p
        class="text-default-700 group-hover:text-primary line-clamp-3 text-sm break-all transition-colors"
      >
        {{ markdownToText(activity.content) }}
      </p>
    </ActivityCardGalgamePreview>
  </ActivityCardShell>
</template>
