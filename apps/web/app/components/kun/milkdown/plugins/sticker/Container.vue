<script setup lang="ts">
import { callCommand } from '@milkdown/kit/utils'
import { insertImageCommand } from '@milkdown/kit/preset/commonmark'
import type { UseEditorReturn } from '@milkdown/vue'

const props = defineProps<{
  editorInfo: UseEditorReturn
}>()

// https://sticker.kungal.com/
const generateStickerArray = () => {
  const result = []
  for (let set = 1; set <= 6; set++) {
    for (let id = 1; id <= 80; id++) {
      result.push(`https://sticker.kungal.com/stickers/KUNgal${set}/${id}.webp`)
    }
  }
  return result.slice(0, -6)
}

const stickerArray = generateStickerArray()

const selectSticker = (stickerUrl: string) => {
  props.editorInfo.get()?.action(
    callCommand(insertImageCommand.key, {
      src: stickerUrl,
      title: 'Sticker',
      alt: 'Sticker'
    })
  )
}
</script>

<template>
  <!-- Every sticker in one vertically-scrolling grid (no pagination). -->
  <div class="grid max-h-72 w-64 grid-cols-5 gap-1 overflow-y-auto p-1">
    <KunImage
      v-for="(sticker, index) in stickerArray"
      :key="index"
      @click="selectSticker(sticker)"
      :src="sticker"
      loading="lazy"
      class="aspect-square w-full cursor-pointer rounded-lg object-contain"
      :alt="`Sticker ${index + 1}`"
      placeholder="/favicon.webp"
    />
  </div>
</template>
