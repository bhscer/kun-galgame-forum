<script setup lang="ts">
import { useWindowScroll } from '@vueuse/core'

withDefaults(
  defineProps<{
    className?: string
  }>(),
  { className: '' }
)

// Fixed offset for the desktop icon rail (w-20); mobile spans full width. The
// bar is permanently pinned to the top — no scroll hide/show.
const offsetClass = 'md:left-[80px] md:w-[calc(100%-88px)]'

// Flat + transparent at the very top; once the page scrolls it gains the surface
// (bg + border + shadow + a real blur), so content sliding underneath reads as
// frosted glass. The old backdrop-blur-[var(--kun-background-blur)] was tied to
// the background-blur SETTING (0 by default) → it never blurred anything.
const { y } = useWindowScroll()
const scrolled = computed(() => y.value > 8)
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
      :class="
        cn(
          'mx-auto flex h-16 w-full max-w-7xl items-center justify-between rounded-b-lg border px-3 transition-all duration-300',
          scrolled
            ? 'bg-content1 border-kun shadow-kun-sm backdrop-blur-md'
            : 'border-transparent backdrop-blur-0'
        )
      "
    >
      <KunTopBarNav />
      <KunTopBarAvatar />
    </div>
  </div>
</template>
