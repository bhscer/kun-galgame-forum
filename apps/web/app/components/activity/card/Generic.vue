<script setup lang="ts">
// Fallback feed card for activity types without a rich card yet: avatar ·
// content · actor + time.
defineProps<{ activity: ActivityItem }>()
</script>

<template>
  <div class="flex items-center gap-3">
    <KunAvatar v-if="activity.actor" :user="activity.actor" />

    <div class="flex flex-col space-y-2">
      <KunLink
        underline="none"
        color="default"
        :to="activity.link"
        class-name="hover:text-primary block space-x-3 break-all transition-colors"
      >
        <KunText
          class-name="whitespace-normal!"
          :content="markdownToText(activity.content)"
        />
      </KunLink>

      <div class="flex items-center space-x-2">
        <span class="text-default-500 text-sm">
          <template v-if="activity.actor"
            >{{ activity.actor.name }} 发布于 </template
          ><KunTime :time="activity.timestamp" />
        </span>
      </div>
    </div>
  </div>
</template>
