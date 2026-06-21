import type { MilkdownPlugin } from '@milkdown/kit/ctx'
import type { Node } from '@milkdown/kit/transformer'
import { $command, $nodeSchema, $remark } from '@milkdown/kit/utils'
import { visit } from 'unist-util-visit'
import type { Node as UnistNode } from 'unist'

// Quote reference (引用楼层). Stored markdown form: `[#<floor>](kungal-reply:<id>)`
// — an ordinary markdown link whose custom `kungal-reply:` scheme the server
// renders into a quote card (no href; hydrated client-side). In the editor it
// becomes a `quote` inline ATOM node carrying the referenced reply id (the
// stable identity) plus a floor-number snapshot for display.
//
// Same shape as the mention plugin: a markdown link is a MARK in milkdown, so a
// $remark transformer rewrites the matching link mdast nodes into a custom
// `quote` node on PARSE (schema only matches its own type, never fighting the
// commonmark link mark), and toMarkdown emits a plain link back. Unlike the
// mention, there is no `@` dropdown — a quote is inserted via insertQuoteCommand
// (wired to a per-floor "引用" button in the reply UI).

export const quoteId = 'quote'
const QUOTE_SCHEME = 'kungal-reply:'

// Minimal shape of the mdast nodes we touch (link on the way in, our synthetic
// `quote` node on the way out).
interface MdastNode extends Node {
  type: string
  url?: string
  value?: string
  replyId?: number
  floor?: number
  children?: MdastNode[]
}

export const quoteSchema = $nodeSchema(quoteId, () => ({
  group: 'inline',
  inline: true,
  atom: true,
  attrs: {
    replyId: { default: 0 },
    floor: { default: 0 }
  },
  parseDOM: [
    {
      // Pasted rendered content (server emits <span class="kun-quote"
      // data-reply-id data-floor>).
      tag: 'span.kun-quote',
      getAttrs: (dom) => {
        const el = dom as HTMLElement
        return {
          replyId: Number.parseInt(el.dataset.replyId ?? '0', 10) || 0,
          floor: Number.parseInt(el.dataset.floor ?? '0', 10) || 0
        }
      }
    }
  ],
  toDOM: (node) => {
    // A non-navigating span chip inside the editor; mirrors the server form so a
    // round-trip through paste/copy is lossless.
    const span = document.createElement('span')
    span.className = 'kun-quote'
    span.dataset.replyId = String(node.attrs.replyId)
    span.dataset.floor = String(node.attrs.floor)
    span.setAttribute('contenteditable', 'false')
    span.textContent = `#${node.attrs.floor}`
    return span
  },
  parseMarkdown: {
    match: (node) => node.type === quoteId,
    runner: (state, node, type) => {
      const n = node as MdastNode
      state.addNode(type, { replyId: n.replyId ?? 0, floor: n.floor ?? 0 })
    }
  },
  toMarkdown: {
    match: (node) => node.type.name === quoteId,
    runner: (state, node) => {
      state.openNode('link', undefined, {
        url: `${QUOTE_SCHEME}${node.attrs.replyId}`
      })
      state.addNode('text', undefined, `#${node.attrs.floor}`)
      state.closeNode()
    }
  }
}))

export const remarkQuotePlugin = $remark('remarkQuote', () => () => {
  const transformer = (tree: UnistNode) => {
    visit(tree, 'link', (node: MdastNode, index, parent: MdastNode) => {
      if (typeof node.url !== 'string' || !node.url.startsWith(QUOTE_SCHEME)) {
        return
      }
      const replyId = Number.parseInt(node.url.slice(QUOTE_SCHEME.length), 10)
      if (!Number.isInteger(replyId) || replyId <= 0) {
        return
      }
      const first = node.children?.[0]
      const floor =
        Number.parseInt(
          (typeof first?.value === 'string' ? first.value : '').replace(
            /^#/,
            ''
          ),
          10
        ) || 0
      if (typeof index === 'number' && parent.children) {
        parent.children.splice(index, 1, {
          type: quoteId,
          replyId,
          floor
        } as MdastNode)
      }
    })
  }
  return transformer
})

// Insert a quote chip at the cursor, replacing the selection. Payload carries
// the referenced reply id + its floor number. Wired to the reply UI's per-floor
// "引用" button (Phase 3); a trailing space lets the caret continue past the
// atom naturally.
export const insertQuoteCommand = $command(
  'InsertKunQuote',
  (ctx) =>
    (payload?: { replyId: number; floor: number }) =>
    (state, dispatch) => {
      if (!payload || !dispatch) {
        return false
      }
      const { replyId, floor } = payload
      if (!Number.isInteger(replyId) || replyId <= 0) {
        return false
      }
      const node = quoteSchema.type(ctx).create({ replyId, floor })
      if (!node) {
        return false
      }
      const tr = state.tr.replaceSelectionWith(node)
      tr.insertText(' ')
      dispatch(tr.scrollIntoView())
      return true
    }
)

export const kunQuotePlugin: MilkdownPlugin[] = [
  quoteSchema,
  remarkQuotePlugin,
  insertQuoteCommand
].flat()
