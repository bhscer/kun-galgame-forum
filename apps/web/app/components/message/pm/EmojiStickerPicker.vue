<script setup lang="ts">
// Chat emoji + sticker picker. Emits `emoji` (inserted at the textarea caret by
// the parent) and `sticker` (sent immediately as its own message). Reuses the
// emoji data + sticker URL scheme the milkdown editor plugins already use, so
// there's a single source of truth for both. The segmented pill toggle is a
// deliberately cleaner take than kun-galgame-patch's underlined tab.
import { emojiArray } from '../../kun/milkdown/plugins/emoji/_isoEmoji'

const emit = defineEmits<{
  emoji: [emoji: string]
  sticker: [url: string]
}>()

const tab = ref<'emoji' | 'sticker'>('emoji')

// sticker.kungal.com — sets KUNgal1..6, ids 1..80 (the last set is 6 short).
// Mirrors the milkdown sticker plugin's generator.
const stickerArray = (() => {
  const result: string[] = []
  for (let set = 1; set <= 6; set++) {
    for (let id = 1; id <= 80; id++) {
      result.push(`https://sticker.kungal.com/stickers/KUNgal${set}/${id}.webp`)
    }
  }
  return result.slice(0, -6)
})()
</script>

<template>
  <div class="w-72 p-2 sm:w-80">
    <!-- Segmented pill toggle (vs. an underlined tab). -->
    <div class="bg-default-100 mb-2 flex rounded-full p-1 text-sm">
      <button
        type="button"
        @click="tab = 'emoji'"
        :class="
          cn(
            'flex-1 rounded-full py-1.5 transition-colors',
            tab === 'emoji'
              ? 'bg-background text-primary font-medium shadow-sm'
              : 'text-default-500 hover:text-default-700'
          )
        "
      >
        表情
      </button>
      <button
        type="button"
        @click="tab = 'sticker'"
        :class="
          cn(
            'flex-1 rounded-full py-1.5 transition-colors',
            tab === 'sticker'
              ? 'bg-background text-primary font-medium shadow-sm'
              : 'text-default-500 hover:text-default-700'
          )
        "
      >
        贴纸
      </button>
    </div>

    <div
      v-show="tab === 'emoji'"
      class="scrollbar-hide grid h-56 grid-cols-8 gap-0.5 overflow-y-auto"
    >
      <button
        v-for="(e, i) in emojiArray"
        :key="i"
        type="button"
        @click="emit('emoji', e)"
        class="hover:bg-default-100 flex aspect-square items-center justify-center rounded-md text-xl"
      >
        {{ e }}
      </button>
    </div>

    <div
      v-show="tab === 'sticker'"
      class="scrollbar-hide grid h-56 grid-cols-4 gap-1 overflow-y-auto"
    >
      <button
        v-for="url in stickerArray"
        :key="url"
        type="button"
        @click="emit('sticker', url)"
        class="hover:bg-default-100 rounded-md p-1"
      >
        <img
          :src="url"
          alt="sticker"
          loading="lazy"
          class="aspect-square w-full object-contain"
        />
      </button>
    </div>
  </div>
</template>
