<script setup lang="ts">
withDefaults(
  defineProps<{
    className?: string
  }>(),
  { className: '' }
)

const isVisible = ref(true)
let lastScrollY = 0

// Hysteresis: only flip visibility on a DELIBERATE scroll (> threshold). Without
// it, a 1px momentum / trackpad micro-reversal flipped isVisible every throttle
// tick, restarting the 300ms transition — the bar (and the area it covers) read
// as a few-px jitter near the top. The layout itself never moves (content keeps
// a constant 76px offset); this is purely the bar stuttering in/out.
const SCROLL_THRESHOLD = 8
const handleScroll = throttle(() => {
  const y = window.scrollY
  if (y < 50) {
    isVisible.value = true
    lastScrollY = y
    return
  }
  const delta = y - lastScrollY
  if (Math.abs(delta) < SCROLL_THRESHOLD) return
  isVisible.value = delta < 0
  lastScrollY = y
}, 100)

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

const { showKUNGalgameSidebarCollapsed } = storeToRefs(
  usePersistSettingsStore()
)

const offsetClass = computed(() =>
  showKUNGalgameSidebarCollapsed.value
    ? 'md:left-[80px] md:w-[calc(100%-88px)]'
    : 'md:left-[260px] md:w-[calc(100%-268px)]'
)
</script>

<template>
  <div
    :class="
      cn(
        'fixed top-0 z-30 mb-3 ml-0 shrink-0 px-1 transition-all duration-300 will-change-transform',
        'left-0 w-full',
        offsetClass,
        isVisible ? 'translate-y-0' : '-translate-y-full',
        className
      )
    "
  >
    <div
      class="bg-background border-default/20 mx-auto flex h-16 w-full max-w-7xl items-center justify-between rounded-b-lg border px-3 backdrop-blur-[var(--kun-background-blur)]"
    >
      <KunTopBarNav />
      <KunTopBarAvatar />
    </div>
  </div>
</template>
