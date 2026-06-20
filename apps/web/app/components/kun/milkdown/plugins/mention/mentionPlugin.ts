import type { MilkdownPlugin } from '@milkdown/kit/ctx'
import type { Node } from '@milkdown/kit/transformer'
import { $nodeSchema, $remark } from '@milkdown/kit/utils'
import { visit } from 'unist-util-visit'
import type { Node as UnistNode } from 'unist'

// @mention. Stored markdown form: `[@name](kungal-user:<id>)` — an ordinary
// markdown link whose custom `kungal-user:` scheme the server renders into a
// mention chip (and parses for notifications). In the editor it becomes a
// `mention` inline ATOM node carrying the user id (the stable identity) plus a
// display-name snapshot.
//
// A markdown link is a MARK (not a node) in milkdown, so a node schema that
// matched links would fight the commonmark link mark. Instead — like the spoiler
// plugin — a $remark transformer rewrites the matching link mdast nodes into a
// custom `mention` mdast node on PARSE, so the schema only ever matches its own
// type. Serialize goes the other way: toMarkdown emits a plain link mdast node,
// which remark-stringify turns back into `[@name](kungal-user:id)`. ($remark
// transformers run on parse only, not on stringify, so the emitted link is left
// alone — same as the spoiler's `||` round-trip.)

export const mentionId = 'mention'
const MENTION_SCHEME = 'kungal-user:'

// Minimal shape of the mdast nodes we touch (link on the way in, our synthetic
// `mention` node on the way out).
interface MdastNode extends Node {
  type: string
  url?: string
  value?: string
  userId?: number
  name?: string
  children?: MdastNode[]
}

export const mentionSchema = $nodeSchema(mentionId, () => ({
  group: 'inline',
  inline: true,
  atom: true,
  attrs: {
    userId: { default: 0 },
    name: { default: '' }
  },
  parseDOM: [
    {
      // Pasted rendered content (server emits <a class="kun-mention" data-uid>).
      tag: 'a.kun-mention',
      getAttrs: (dom) => {
        const el = dom as HTMLElement
        return {
          userId: Number.parseInt(el.dataset.uid ?? '0', 10) || 0,
          name: (el.textContent ?? '').replace(/^@/, '')
        }
      }
    }
  ],
  toDOM: (node) => {
    // A non-navigating span chip inside the editor (a real <a> would steal the
    // click and leave the page mid-compose). data-uid mirrors the server form.
    const span = document.createElement('span')
    span.className = 'kun-mention'
    span.dataset.uid = String(node.attrs.userId)
    span.setAttribute('contenteditable', 'false')
    span.textContent = `@${node.attrs.name}`
    return span
  },
  parseMarkdown: {
    match: (node) => node.type === mentionId,
    runner: (state, node, type) => {
      const n = node as MdastNode
      state.addNode(type, { userId: n.userId ?? 0, name: n.name ?? '' })
    }
  },
  toMarkdown: {
    match: (node) => node.type.name === mentionId,
    runner: (state, node) => {
      state.openNode('link', undefined, {
        url: `${MENTION_SCHEME}${node.attrs.userId}`
      })
      state.addNode('text', undefined, `@${node.attrs.name}`)
      state.closeNode()
    }
  }
}))

export const remarkMentionPlugin = $remark('remarkMention', () => () => {
  const transformer = (tree: UnistNode) => {
    visit(tree, 'link', (node: MdastNode, index, parent: MdastNode) => {
      if (
        typeof node.url !== 'string' ||
        !node.url.startsWith(MENTION_SCHEME)
      ) {
        return
      }
      const userId = Number.parseInt(node.url.slice(MENTION_SCHEME.length), 10)
      if (!Number.isInteger(userId) || userId <= 0) {
        return
      }
      const first = node.children?.[0]
      const name = (
        typeof first?.value === 'string' ? first.value : ''
      ).replace(/^@/, '')
      if (typeof index === 'number' && parent.children) {
        parent.children.splice(index, 1, {
          type: mentionId,
          userId,
          name
        } as MdastNode)
      }
    })
  }
  return transformer
})

export const kunMentionPlugin: MilkdownPlugin[] = [
  mentionSchema,
  remarkMentionPlugin
].flat()
