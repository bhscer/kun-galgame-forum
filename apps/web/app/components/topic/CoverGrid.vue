<script setup lang="ts">
// Renders a topic's covers as a 九宫格-style grid. Each entry is a
// /image/<hash> content token. Purely visual — the caller wraps it in a link.
const props = defineProps<{
  images: string[]
}>()

// Cap defensively at the stored max. 1 = single banner; 2/4 = 2-col; else 3-col.
const shown = computed(() => props.images.slice(0, 9))
const isSingle = computed(() => shown.value.length === 1)
const gridClass = computed(() => {
  const n = shown.value.length
  if (n === 1) return 'grid-cols-1'
  if (n === 2 || n === 4) return 'grid-cols-2'
  return 'grid-cols-3'
})
</script>

<template>
  <div v-if="shown.length" :class="cn('grid w-full gap-1.5', gridClass)">
    <div
      v-for="(token, idx) in shown"
      :key="`${idx}-${token}`"
      :class="
        cn(
          'bg-default-100 overflow-hidden rounded-lg',
          // Capped heights so a wide desktop column doesn't make covers huge —
          // a single banner stays ~256px, grid cells ~160px (community-feed sizing).
          isSingle ? 'h-64' : 'h-40'
        )
      "
    >
      <!-- Plain <img> on the ABSOLUTE CDN URL (imageTokenUrl resolves the
           /image/<hash> token): skips @nuxt/image IPX (which 404s on the token)
           AND the /image 302 hop — straight to the CDN's optimized webp. -->
      <img
        :src="imageTokenUrl(token)"
        alt="话题封面"
        loading="lazy"
        class="h-full w-full object-cover"
      />
    </div>
  </div>
</template>
