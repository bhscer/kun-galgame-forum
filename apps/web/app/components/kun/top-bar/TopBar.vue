<script setup lang="ts">
import { useWindowScroll } from '@vueuse/core'

withDefaults(
  defineProps<{
    className?: string
  }>(),
  { className: '' }
)

// Fixed offset for the desktop icon rail (w-20); mobile spans full width.
const offsetClass = 'md:left-[80px] md:w-[calc(100%-88px)]'

// Flat + edge-to-edge (px-0) + transparent at the very top; once the page scrolls
// it eases to the inset surface (px-3 + bg + border + shadow). The blur is kept
// CONSTANT (not animated) and the bar sits on its own GPU layer (transform-gpu),
// so crossing the threshold no longer animates backdrop-filter — that animation
// was what stuttered the page. At the top there's nothing behind the bar
// (the page's pt clears it), so the constant blur is invisible.
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
          'mx-auto flex h-16 w-full max-w-7xl transform-gpu items-center justify-between rounded-b-lg border backdrop-blur-md transition-all duration-200',
          scrolled
            ? 'bg-content1 border-kun px-3 shadow-kun-sm'
            : 'border-transparent bg-transparent px-0 shadow-none'
        )
      "
    >
      <KunTopBarNav />
      <KunTopBarAvatar />
    </div>
  </div>
</template>
