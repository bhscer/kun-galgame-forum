<script setup lang="ts">
import { kunSanitize } from '@kun/ui/utils/sanitize'

const props = defineProps<{
  message: MessageSystemMessage
}>()

// Defense-in-depth: admin/system broadcast content is rendered as HTML, so
// sanitize before binding to v-html (consistent with SnapshotDiff.vue and
// moyu's message sinks). Even though the write path is admin-only today, an
// unsanitized v-html here would be a stored-XSS the moment any user-influenced
// text reaches a system message.
const sanitizedContent = computed(() =>
  kunSanitize(props.message.content['zh-cn'] ?? '')
)
</script>

<template>
  <div
    class="space-y-3 rounded-lg p-2"
    :class="message.isRead ? 'message-read' : ''"
  >
    <div class="flex items-center break-all">
      <div class="ml-2 text-lg">
        <KunIcon
          class="text-secondary"
          v-if="!message.isRead"
          name="lucide:info"
        />
        <KunIcon
          class="text-default"
          v-if="message.isRead"
          name="lucide:check-check"
        />
      </div>
      <KunAvatar :disable-floating="true" :user="message.admin" />
      <span class="text-default-500 text-sm">
        {{ formatTimeDifference(message.created) }}
      </span>
    </div>

    <div class="leading-8 break-all" v-html="sanitizedContent" />
  </div>
</template>
