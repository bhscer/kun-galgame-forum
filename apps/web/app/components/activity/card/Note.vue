<script setup lang="ts">
// TODO_CREATION / UPDATE_LOG_CREATION — the site's 待办 / 更新日志 entries (both
// link to /update). A small typed label (icon + name) over the content — nicer
// than the bare generic fallback these used to fall through to.
const props = defineProps<{ activity: ActivityItem }>()

const meta = computed(() =>
  props.activity.type === 'TODO_CREATION'
    ? { icon: 'lucide:list-checks', label: '待办事项', color: 'text-warning-600' }
    : { icon: 'lucide:megaphone', label: '更新日志', color: 'text-primary' }
)
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <KunLink
      underline="none"
      color="default"
      :to="activity.link"
      class-name="group block space-y-1.5"
    >
      <span
        :class="cn('flex items-center gap-1.5 text-sm font-medium', meta.color)"
      >
        <KunIcon :name="meta.icon" class="size-4 shrink-0" />
        {{ meta.label }}
      </span>
      <p
        class="group-hover:text-primary line-clamp-4 text-base break-all transition-colors"
      >
        {{ markdownToText(activity.content) }}
      </p>
    </KunLink>
  </ActivityCardShell>
</template>
