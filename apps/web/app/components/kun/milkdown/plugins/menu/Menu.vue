<script setup lang="ts">
import { callCommand } from '@milkdown/kit/utils'
import { insertImageCommand } from '@milkdown/kit/preset/commonmark'
import { commands } from './_buttonList'
import { tabs, activeTab } from '../../atom'
import type { UseEditorReturn } from '@milkdown/vue'
import type { CmdKey } from '@milkdown/kit/core'

const props = defineProps<{
  editorInfo: UseEditorReturn
  // Whether this editor permits images. Gates BOTH the upload-image button and
  // the sticker inserter (stickers are images). false → text-only toolbar
  // (still keeps unicode emoji, which are plain text). The galgame 简介 editor
  // passes false.
  allowImage: boolean
}>()

const { get } = props.editorInfo
const input = ref<HTMLElement>()

const call = <T,>(command: CmdKey<T>, payload?: T, callback?: () => void) => {
  const result = get()?.action(callCommand(command, payload))
  if (callback) {
    callback()
  }
  return result
}

const handleFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files || !input.files[0]) {
    return
  }
  const file = input.files[0]

  const formData = new FormData()
  formData.append('image', file)

  useMessage(10107, 'info')
  const result = await kunFetch<string>('/image/topic', {
    method: 'POST',
    body: formData,
    watch: false
  })

  if (result) {
    const filename = file.name.replace(/[^a-zA-Z0-9 ]/g, '')
    const userName = usePersistUserStore().name
    const imageName = `${userName}-${Date.now()}-${filename}`
    call(insertImageCommand.key, {
      src: result ?? '',
      title: imageName,
      alt: imageName
    })
    useMessage(10108, 'success')
  }
}

const toggleTab = () => {
  activeTab.value = activeTab.value === 'preview' ? 'code' : 'preview'
}

const currentTabLabel = computed(() => {
  return activeTab.value === 'preview' ? tabs[1]!.textValue : tabs[0]!.textValue
})

// Merged sticker/emoji popover tab. Default 贴纸; the 贴纸 tab only shows when the
// editor allows images (stickers ARE images) — the text-only editor shows 表情 alone.
const insertTab = ref('sticker')
const insertTabItems = [
  { value: 'sticker', textValue: '贴纸', icon: 'lucide:sticker' },
  { value: 'emoji', textValue: '表情', icon: 'lucide:smile-plus' }
]
</script>

<template>
  <div class="flex flex-wrap items-center space-x-1">
    <KunButton variant="light" @click="toggleTab">
      {{ currentTabLabel }}
    </KunButton>

    <template v-if="activeTab === 'preview'">
      <KunMilkdownPluginsHeader :editor-info="editorInfo" />

      <KunButton
        :is-icon-only="true"
        v-for="(btn, index) in commands"
        :key="index"
        class-name="text-xl"
        variant="light"
        @click="call(btn.command.key, btn.payload)"
      >
        <KunIcon class="text-foreground" :name="btn.icon" />
      </KunButton>

      <KunButton
        :is-icon-only="true"
        v-if="props.allowImage"
        variant="light"
        class-name="text-xl"
        @click="input?.click()"
      >
        <KunIcon class="text-foreground" name="lucide:image-plus" />
        <input
          hidden
          ref="input"
          type="file"
          accept=".jpg, .jpeg, .png, .webp"
          @change="handleFileChange($event)"
        />
      </KunButton>

      <!-- One merged popover for stickers + emoji. Emoji icon as the trigger; a
           贴纸/表情 tab switches the body (tab hidden when images are disallowed,
           leaving 表情 only). -->
      <!-- kun-ui 2.3 auto-positions (flip/shift/size-cap) by default, so the old
           manual `-left-28` nudge is gone — the panel now fits the edge on its own. -->
      <KunPopover opaque>
        <template #trigger>
          <KunButton variant="light" class-name="text-xl" :is-icon-only="true">
            <KunIcon class="text-foreground" name="lucide:smile-plus" />
          </KunButton>
        </template>

        <div class="flex flex-col gap-1.5">
          <KunTab
            v-if="props.allowImage"
            v-model="insertTab"
            :items="insertTabItems"
            variant="underlined"
            color="primary"
            full-width
          />
          <KunMilkdownPluginsStickerContainer
            v-if="props.allowImage && insertTab === 'sticker'"
            :editor-info="editorInfo"
          />
          <KunMilkdownPluginsEmojiContainer
            v-if="!props.allowImage || insertTab === 'emoji'"
            :editor-info="editorInfo"
          />
        </div>
      </KunPopover>
    </template>
  </div>
</template>
