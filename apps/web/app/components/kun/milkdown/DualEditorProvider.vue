<script setup lang="ts">
import { MilkdownProvider } from '@milkdown/vue'
import { ProsemirrorAdapterProvider } from '@prosemirror-adapter/vue'
import { activeTab } from './atom'

import type { ReplyReference } from '~/store/types/topic/reply'

defineProps<{
  valueMarkdown: string
  language?: Language
  // Forwarded to the editor: true removes all image insertion (upload button,
  // paste/drop, sticker). The galgame 简介 editor sets it.
  disableImage?: boolean
  // Forwarded 「引用」 signal (reply editor only).
  pendingQuote?: ReplyReference | null
}>()

const emits = defineEmits<{
  setMarkdown: [value: string]
  quoteInserted: []
}>()

const cmAPI = ref({
  update: (_: string) => {}
})

const saveMarkdown = (markdown: string) => {
  cmAPI.value.update(markdown)
  emits('setMarkdown', markdown)
}

const setCmAPI = (api: { update: (markdown: string) => void }) => {
  cmAPI.value = api
}
</script>

<template>
  <MilkdownProvider>
    <ProsemirrorAdapterProvider>
      <div class="space-y-3">
        <KunMilkdownEditor
          :value-markdown="valueMarkdown"
          @save-markdown="saveMarkdown"
          :language="language ?? 'zh-cn'"
          :disable-image="disableImage"
          :pending-quote="pendingQuote"
          @quote-inserted="emits('quoteInserted')"
        >
          <template #footer>
            <slot />
          </template>
        </KunMilkdownEditor>

        <template v-if="activeTab === 'code'">
          <KunMilkdownCodemirror
            :markdown="valueMarkdown"
            @set-cm-api="setCmAPI"
            @on-change="(value) => emits('setMarkdown', value)"
          />
        </template>
      </div>
    </ProsemirrorAdapterProvider>
  </MilkdownProvider>
</template>
