<script setup lang="ts">
withDefaults(
  defineProps<{
    className?: string
  }>(),
  { className: '' }
)

// Fixed offset for the desktop icon rail (w-20); mobile spans full width. The
// bar is permanently pinned to the top — no scroll hide/show.
const offsetClass = 'md:left-[80px] md:w-[calc(100%-88px)]'
</script>

<template>
  <div
    :class="
      cn(
        'fixed top-0 z-30 mb-3 ml-0 shrink-0 px-1',
        'left-0 w-full',
        offsetClass,
        className
      )
    "
  >
    <div
      class="bg-content1 border-kun mx-auto flex h-16 w-full max-w-7xl items-center justify-between rounded-b-lg border px-3 shadow-kun-sm backdrop-blur-[var(--kun-background-blur)]"
    >
      <KunTopBarNav />
      <KunTopBarAvatar />
    </div>
  </div>
</template>
