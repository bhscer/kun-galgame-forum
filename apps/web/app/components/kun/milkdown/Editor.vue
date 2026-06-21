<script setup lang="ts">
// Milkdown core
import {
  Editor,
  rootCtx,
  defaultValueCtx,
  editorViewCtx
} from '@milkdown/kit/core'
import { Selection } from '@milkdown/prose/state'
import type { EditorView as PmEditorView } from '@milkdown/prose/view'
import { Milkdown, useEditor } from '@milkdown/vue'
import { commonmark } from '@milkdown/kit/preset/commonmark'
import { gfm } from '@milkdown/kit/preset/gfm'
// Milkdown Plugins
import { history } from '@milkdown/kit/plugin/history'
import { listener, listenerCtx } from '@milkdown/kit/plugin/listener'
import { clipboard } from '@milkdown/kit/plugin/clipboard'
import { indent } from '@milkdown/kit/plugin/indent'
import { trailing } from '@milkdown/kit/plugin/trailing'
import { usePluginViewFactory } from '@prosemirror-adapter/vue'
import {
  upload,
  uploadConfig,
  type Uploader
} from '@milkdown/kit/plugin/upload'

// Custom plugins
import { activeTab } from './atom'
import { kunUploader, kunUploadWidgetFactory } from './plugins/upload/uploader'
import { tooltipFactory } from '@milkdown/kit/plugin/tooltip'
import Tooltip from './plugins/tooltip/Tooltip.vue'
import { slashFactory } from '@milkdown/kit/plugin/slash'
import MentionDropdown from './plugins/mention/MentionDropdown.vue'
import { replaceAll } from '@milkdown/kit/utils'
import {
  stopLinkCommand,
  linkCustomKeymap
} from './plugins/stop-link/stopLinkPlugin'
import { kunSpoilerPlugin } from './plugins/spoiler/spoilerPlugin'
import { kunMentionPlugin } from './plugins/mention/mentionPlugin'
import { kunQuotePlugin } from './plugins/quote/quotePlugin'

// Code Block
import { defaultKeymap, indentWithTab } from '@codemirror/commands'
import { keymap, EditorView } from '@codemirror/view'
import {
  codeBlockComponent,
  codeBlockConfig,
  type CodeBlockConfig
} from '@milkdown/kit/component/code-block'
import { basicSetup } from 'codemirror'
import {
  chevronDownIcon,
  clearIcon,
  copyIcon,
  editIcon,
  searchIcon,
  visibilityOffIcon
} from './plugins/code/icons'
import { languages } from '@codemirror/language-data'
import { kunCM } from './codemirror/theme'

// katex
import { blockKatexSchema } from './plugins/katex/blockKatex'
import { mathInlineSchema } from './plugins/katex/inlineKatex'
import { toggleLatexCommand } from './plugins/katex/command'
import {
  mathBlockInputRule,
  mathInlineInputRule
} from './plugins/katex/inputRule'
import { remarkMathBlockPlugin, remarkMathPlugin } from './plugins/katex/remark'
import katex from 'katex'
import type { KatexOptions } from 'katex'

const props = defineProps<{
  valueMarkdown: string
  language: Language
  // When true, all image-insertion paths are removed: the toolbar upload
  // button, paste/drop upload, AND the sticker inserter (stickers render as
  // images). Used by the galgame 简介 editor — intro images now live in the
  // wiki gallery, so the introduction itself must stay image-free.
  disableImage?: boolean
}>()

const emits = defineEmits<{
  saveMarkdown: [markdown: string]
}>()

const valueMarkdown = computed(() => props.valueMarkdown)

const tooltip = tooltipFactory('Text')
// @mention autocomplete: @ trigger → debounced /user/search → insert mention node.
const mentionSlash = slashFactory('kunMention')
const pluginViewFactory = usePluginViewFactory()
const container = ref<HTMLElement | null>(null)
const toolbar = ref<HTMLElement | null>(null)
const editorContent = ref('')
// The markdown the editor last emitted; lets the valueMarkdown watch below
// distinguish an EXTERNAL change to the model (the 「引用」 button appending a
// token, switching to edit content) from the editor's own edits, so only the
// former triggers a replaceAll (the latter would reset the cursor every keystroke).
let lastEmitted = props.valueMarkdown

const renderLatex = (content: string, options?: KatexOptions) => {
  const html = katex.renderToString(content, {
    ...options,
    throwOnError: false,
    displayMode: true
  })
  return html
}

// When a 「引用」 insert leaves the doc ending in a blockquote header
// (`> 回复 @ #`), append an empty paragraph after it and put the caret there, so
// the reply body is typed on the line BELOW the header (the user can Backspace
// to merge). Narrowly gated on a trailing blockquote so normal content / other
// editors are untouched. Called on mount (panel opened fresh by 「引用」) AND from
// the external-sync watch (panel already open, token appended live).
const placeCaretBelowHeader = (view: PmEditorView) => {
  const last = view.state.doc.lastChild
  if (!last || last.type.name !== 'blockquote') {
    return
  }
  const para = view.state.schema.nodes.paragraph?.createAndFill()
  if (!para) {
    return
  }
  const tr = view.state.tr.insert(view.state.doc.content.size, para)
  view.dispatch(tr.setSelection(Selection.atEnd(tr.doc)).scrollIntoView())
}

const editorInfo = useEditor((root) => {
  const editor = Editor.make()
    .config((ctx) => {
      ctx.set(rootCtx, root)
      ctx.set(defaultValueCtx, valueMarkdown.value)

      const listener = ctx.get(listenerCtx)
      listener.markdownUpdated((ctx, markdown, prevMarkdown) => {
        if (markdown !== prevMarkdown) {
          editorContent.value = markdown
          // Record what the editor produced so the valueMarkdown watch below can
          // tell an external change (e.g. the 「引用」 button appending a token)
          // from the editor's own edits and avoid a replaceAll feedback loop.
          lastEmitted = markdown
          emits('saveMarkdown', markdown)
        }
      })
      // Fresh-mount case: the panel opened by 「引用」 mounts with the header in
      // defaultValueCtx (no valueMarkdown change → the watch won't fire), so
      // place the caret below the header here.
      listener.mounted((ctx) => {
        placeCaretBelowHeader(ctx.get(editorViewCtx))
      })

      // Only wire the uploader when images are allowed. uploadConfig.key is
      // registered by the `upload` plugin, which we omit below when
      // disableImage is set — touching the key without the plugin would throw.
      if (!props.disableImage) {
        ctx.update(uploadConfig.key, (prev) => ({
          ...prev,
          // kunUploader is typed against @milkdown/plugin-upload's Uploader, whose
          // Ctx resolves to a different (duplicated) @milkdown/ctx copy than kit's
          // re-export. The two Ctx types are runtime-identical but nominally
          // distinct (private brand), so go through `unknown` to unify them.
          uploader: kunUploader as unknown as Uploader,
          uploadWidgetFactory: kunUploadWidgetFactory
        }))
      }

      ctx.set(tooltip.key, {
        view: pluginViewFactory({
          component: Tooltip
        })
      })

      ctx.set(mentionSlash.key, {
        view: pluginViewFactory({
          component: MentionDropdown
        })
      })

      const extensions = [
        kunCM(),
        EditorView.lineWrapping,
        keymap.of(defaultKeymap.concat(indentWithTab)),
        basicSetup
      ]
      // if (theme) {
      //   extensions.push(theme)
      // }

      ctx.update(codeBlockConfig.key, (defaultConfig) => ({
        extensions,
        languages,
        expandIcon: chevronDownIcon,
        searchIcon: searchIcon,
        clearSearchIcon: clearIcon,
        searchPlaceholder: '搜索咒文',
        copyText: '复制咒文',
        copyIcon: copyIcon,
        onCopy: () => {},
        noResultText: '无结果',
        renderLanguage: defaultConfig.renderLanguage,
        previewLoading: '加载中...',
        renderPreview: defaultConfig.renderPreview,
        previewToggleButton: (previewOnlyMode) => {
          const icon = previewOnlyMode ? editIcon : visibilityOffIcon
          const text = previewOnlyMode ? '编辑' : '隐藏'
          return [icon, text].map((v) => v.trim()).join(' ')
        },
        previewLabel: defaultConfig.previewLabel
        // previewLoading: config.previewLoading || defaultConfig.previewLoading,
        // previewOnlyByDefault:
        //   config.previewOnlyByDefault ?? defaultConfig.previewOnlyByDefault
      }))

      const katexOptions: KatexOptions = {}

      ctx.update(codeBlockConfig.key, (prev) => ({
        ...prev,
        renderPreview: (language, content, applyPreview) => {
          if (language.toLowerCase() === 'latex' && content.length > 0) {
            return renderLatex(content, katexOptions)
          }
          const renderPreview = prev.renderPreview
          return renderPreview(language, content, applyPreview)
        }
      }))
    })
    .use(history)
    .use(commonmark)
    .use(gfm)
    .use(listener)
    .use(clipboard)
    .use(indent)
    .use(trailing)
    .use(tooltip)
    .use(mentionSlash)
    .use(codeBlockComponent)
    .use([kunSpoilerPlugin, stopLinkCommand, linkCustomKeymap].flat())
    .use(kunMentionPlugin)
    .use(kunQuotePlugin)
    .use(remarkMathPlugin)
    .use(remarkMathBlockPlugin)
    .use(mathInlineSchema)
    .use(mathInlineInputRule)
    .use(mathBlockInputRule)
    .use(blockKatexSchema)
    .use(toggleLatexCommand)

  // Image upload (toolbar + paste/drop) is opt-out: omit the `upload` plugin
  // entirely for image-free editors (galgame 简介). The toolbar button and the
  // sticker inserter are hidden separately via the menu's allow-image prop.
  if (!props.disableImage) {
    editor.use(upload)
  }

  return editor
})

const textCount = computed(() => markdownToText(props.valueMarkdown).length)

watch(
  () => [props.language],
  () => {
    editorInfo.get()?.action(replaceAll(valueMarkdown.value))
  }
)

// Re-sync the doc when valueMarkdown changes from OUTSIDE the editor — e.g. the
// 「引用」 button appends an @mention/#quote token to the draft. Guarded by
// lastEmitted so the editor's own edits (which round-trip back through the model)
// don't trigger a replaceAll and reset the cursor on every keystroke.
watch(
  () => props.valueMarkdown,
  (val) => {
    if (val === lastEmitted) {
      return
    }
    lastEmitted = val
    const editor = editorInfo.get()
    if (!editor) {
      return
    }
    editor.action(replaceAll(val))
    // Live-append case (panel already open): a 「引用」 token was appended to the
    // draft — drop the caret on the line below the new header.
    editor.action((ctx) => {
      placeCaretBelowHeader(ctx.get(editorViewCtx))
    })
  }
)
</script>

<template>
  <div ref="container" class="space-y-3">
    <KunMilkdownPluginsMenu
      ref="toolbar"
      :editor-info="editorInfo"
      :allow-image="!disableImage"
    />

    <template v-if="activeTab === 'preview'">
      <Milkdown />

      <div class="flex items-center justify-between text-sm">
        <slot name="footer" />

        <div class="flex shrink-0 items-center gap-2">
          <KunChip color="success">
            <KunIcon
              name="simple-icons:markdown"
              class="text-success-700 dark:text-success"
            />
            Markdown 支持
          </KunChip>
          <span>
            {{ `${textCount} 字` }}
          </span>
        </div>
      </div>

      <div class="text-default-500 text-sm">
        特殊语法: 您可以使用 ||隐藏文本|| 来隐藏图片或者文字 (目前依然禁止 R18
        图片内容)
      </div>
    </template>
  </div>
</template>
