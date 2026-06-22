<script setup lang="ts">
// Renders a topic's 1..9 feed-card covers as a 九宫格-style grid. Each entry is a
// /image/<hash> content token, usable directly as an <img src>. Purely visual —
// the caller wraps it in the card's <KunLink>, so a click opens the topic.
const props = defineProps<{
  images: string[]
}>()

// Cap defensively at the stored max. 1 = single wide banner; 2/4 = 2-col;
// everything else = 3-col (the classic feed grid).
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
      class="bg-default-100 overflow-hidden rounded-lg"
    >
      <KunImage
        :src="token"
        alt="话题封面"
        loading="lazy"
        :class-name="
          cn('w-full object-cover', isSingle ? 'aspect-video' : 'aspect-square')
        "
      />
    </div>
  </div>
</template>
