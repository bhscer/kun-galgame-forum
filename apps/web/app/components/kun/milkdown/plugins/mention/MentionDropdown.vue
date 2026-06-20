<script setup lang="ts">
import { SlashProvider } from '@milkdown/kit/plugin/slash'
import { usePluginViewContext } from '@prosemirror-adapter/vue'
import { TextSelection } from '@milkdown/prose/state'
import type { EditorView } from '@milkdown/prose/view'
import type { VNodeRef } from 'vue'
import { mentionId } from './mentionPlugin'

// One row of the @mention autocomplete — mirrors the Go `dto.MentionUser`.
interface MentionUser {
  id: number
  name: string
  avatar: string
}

// The @query token at the cursor: `@` followed by up to 30 non-space,
// non-`@` chars, anchored to start-of-text or a space so `foo@bar` (emails)
// never triggers. Group 1 is the query (without the `@`).
const MENTION_RE = /(?:^|\s)@([^\s@]{0,30})$/

const { view, prevState } = usePluginViewContext()

const divRef = ref<VNodeRef>()
let provider: SlashProvider

const visible = ref(false)
const query = ref('')
const results = ref<MentionUser[]>([])
const activeIndex = ref(0)
const searching = ref(false)
// Doc range of the `@query` text, captured in shouldShow so selecting a user
// knows exactly what to replace with the mention node.
let range: { from: number; to: number } | null = null
// Guards a stale async response from overwriting a newer query's results.
let searchSeq = 0

let searchTimer: ReturnType<typeof setTimeout> | null = null

const reset = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
  query.value = ''
  results.value = []
  activeIndex.value = 0
  searching.value = false
  range = null
}

// Debounce the NETWORK call only (not the dropdown), so the @query + UI stay
// responsive per keystroke while OAuth /users/search is hit at most ~5x/sec.
const search = (q: string) => {
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
  if (!q) {
    results.value = []
    searching.value = false
    return
  }
  searching.value = true
  searchTimer = setTimeout(async () => {
    const seq = ++searchSeq
    const data = await kunFetch<MentionUser[]>('/user/search', {
      query: { q, limit: 8 }
    })
    // A newer keystroke already fired — drop this response.
    if (seq !== searchSeq) {
      return
    }
    results.value = data ?? []
    activeIndex.value = 0
    searching.value = false
  }, 180)
}

// The text of the current textblock from its start up to the cursor — computed
// straight from the doc rather than via SlashProvider.getContent (whose default
// only matches `paragraph` blocks and layers on focus checks we don't need).
const queryBeforeCursor = (v: EditorView): string | null => {
  const { selection } = v.state
  if (!selection.empty) {
    return null
  }
  const { $from } = selection
  const before = $from.parent.textBetween(
    Math.max(0, $from.parentOffset - 100),
    $from.parentOffset,
    undefined,
    '￼'
  )
  const match = before.match(MENTION_RE)
  return match ? (match[1] ?? '') : null
}

// SlashProvider calls this on every editor update. It must stay cheap +
// synchronous: detect the @query, capture its range, kick the search off only
// when the query text actually changed, and return whether to show.
const shouldShow = (v: EditorView): boolean => {
  const q = queryBeforeCursor(v)
  if (q === null) {
    return false
  }
  const cursor = v.state.selection.$from.pos
  // `@` + query length — the leading space (if any) is matched but NOT replaced.
  range = { from: cursor - (q.length + 1), to: cursor }
  if (q !== query.value) {
    query.value = q
    search(q)
  }
  return true
}

const selectUser = (user: MentionUser) => {
  const v = view.value
  if (!v || !range) {
    return
  }
  const type = v.state.schema.nodes[mentionId]
  if (!type) {
    return
  }
  const node = type.create({ userId: user.id, name: user.name })
  const { from, to } = range
  const tr = v.state.tr.replaceWith(from, to, node)
  // A trailing space so the caret lands clear of the atom and typing continues
  // naturally instead of nudging the chip.
  const after = from + node.nodeSize
  tr.insertText(' ', after)
  tr.setSelection(TextSelection.create(tr.doc, after + 1))
  v.dispatch(tr.scrollIntoView())
  v.focus()
  provider.hide()
}

// The editor keeps DOM focus while the dropdown is open, so navigation keys go
// to ProseMirror. Intercept them in the capture phase (before PM sees them),
// but only while the dropdown is actually showing suggestions.
const onKeydown = (e: KeyboardEvent) => {
  if (!visible.value || !results.value.length) {
    if (visible.value && e.key === 'Escape') {
      e.preventDefault()
      provider.hide()
    }
    return
  }
  const last = results.value.length - 1
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      e.stopPropagation()
      activeIndex.value = activeIndex.value >= last ? 0 : activeIndex.value + 1
      break
    case 'ArrowUp':
      e.preventDefault()
      e.stopPropagation()
      activeIndex.value = activeIndex.value <= 0 ? last : activeIndex.value - 1
      break
    case 'Enter':
    case 'Tab': {
      const picked = results.value[activeIndex.value]
      if (picked) {
        e.preventDefault()
        e.stopPropagation()
        selectUser(picked)
      }
      break
    }
    case 'Escape':
      e.preventDefault()
      e.stopPropagation()
      provider.hide()
      break
  }
}

onMounted(() => {
  provider = new SlashProvider({
    content: divRef.value as unknown as HTMLElement,
    trigger: '@',
    shouldShow,
    // Run shouldShow on every update (no provider debounce) so the @query and
    // the panel state track each keystroke; the network call is debounced in
    // search() instead. A provider debounce would freeze the panel on a stale
    // query while typing fast.
    debounce: 0
  })
  provider.onShow = () => {
    visible.value = true
  }
  provider.onHide = () => {
    visible.value = false
    reset()
  }
  provider.update(view.value, prevState.value)
  // Capture phase so we beat ProseMirror's own keydown handling.
  window.addEventListener('keydown', onKeydown, true)
})

watch([view, prevState], () => {
  provider?.update(view.value, prevState.value)
})

onBeforeUnmount(() => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  window.removeEventListener('keydown', onKeydown, true)
  provider?.destroy()
})
</script>

<template>
  <!-- Rendered unconditionally (no v-if): the SlashProvider appendChild's this
       element out of Vue's Teleport, so a v-if that later flips would orphan it
       in the DOM and freeze its content. Visibility is the data-show CSS. -->
  <div ref="divRef" class="kun-mention-dropdown" data-show="false">
    <template v-if="results.length">
      <button
        v-for="(u, i) in results"
        :key="u.id"
        type="button"
        :class="
          cn(
            'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left',
            i === activeIndex
              ? 'bg-primary-100 text-primary-700'
              : 'hover:bg-default-100'
          )
        "
        @mousedown.prevent="selectUser(u)"
        @mouseenter="activeIndex = i"
      >
        <img
          v-if="u.avatar"
          :src="u.avatar"
          :alt="u.name"
          class="size-6 shrink-0 rounded-full object-cover"
        />
        <span
          v-else
          class="bg-default-200 text-default-600 flex size-6 shrink-0 items-center justify-center rounded-full text-xs"
        >
          {{ u.name.charAt(0) }}
        </span>
        <span class="truncate text-sm">{{ u.name }}</span>
      </button>
    </template>

    <div v-else-if="searching" class="text-default-400 px-2 py-1.5 text-sm">
      搜索中…
    </div>
    <div v-else-if="query" class="text-default-400 px-2 py-1.5 text-sm">
      无匹配用户
    </div>
    <div v-else class="text-default-400 px-2 py-1.5 text-sm">
      输入用户名以搜索
    </div>
  </div>
</template>
