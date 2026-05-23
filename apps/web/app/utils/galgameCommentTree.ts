import type { SerializeObject } from 'nitropack'

// Flat-tier mutation helpers for the galgame comment tree.
//
// The visual model is intentionally TWO TIERS only: root + a flat list
// of replies. The DB still tracks the full parent_comment_id chain
// (so true "回复给某条" semantics survive via targetUser), but the
// frontend always renders root.replies as a single flat group with no
// further indent. These helpers therefore treat root.replies as a
// flat array — never walk inside a reply looking for sub-replies.
//
// Both Container.vue (inline list, ≤3 replies per root) and
// ThreadDrawer.vue (full thread, unbounded) call into the same logic.
// All returns are immutable: the input root is not mutated, and a new
// root object is returned so Vue can detect the change even when
// stored behind a shallowRef.

type Node = SerializeObject<GalgameComment>

// Prepend `newReply` at the TOP of root.replies and bump replyCount.
// The inline + drawer views both render replies newest-first, so a
// just-posted reply must land at index 0 to be immediately visible
// at the top of the list. Always returns a fresh root object.
export const prependReplyToRoot = (root: Node, newReply: Node): Node => {
  return {
    ...root,
    replies: [newReply, ...(root.replies ?? [])],
    replyCount: (root.replyCount ?? 0) + 1
  }
}

// Remove the reply matching `removedId` and (transitively) any of its
// descendants that happen to be loaded in the flat list. The DB
// cascades the actual rows via parent_comment_id FK; this mirrors that
// on the client so the visible list doesn't dangle.
//
// `removedSize` is the subtree size reported by the caller (the
// deleted node + all its DB descendants — most of which may not even
// be loaded). We decrement replyCount by that exact amount because
// the count is authoritative for the WHOLE thread, not just what's
// visible.
export const removeReplyFromRoot = (
  root: Node,
  removedId: number,
  removedSize: number
): { node: Node; pruned: boolean } => {
  const replies = root.replies ?? []
  if (!replies.some((r) => r.id === removedId)) {
    return { node: root, pruned: false }
  }

  // Collect removedId + any loaded descendants. parentCommentId chains
  // can be multi-hop within the loaded set (a depth-3 reply pointing
  // at a depth-2 reply), so iterate until the frontier stops growing.
  const toRemove = new Set<number>([removedId])
  let added = true
  while (added) {
    added = false
    for (const r of replies) {
      if (
        r.parentCommentId != null &&
        toRemove.has(r.parentCommentId) &&
        !toRemove.has(r.id)
      ) {
        toRemove.add(r.id)
        added = true
      }
    }
  }

  return {
    node: {
      ...root,
      replies: replies.filter((r) => !toRemove.has(r.id)),
      replyCount: Math.max(0, (root.replyCount ?? 0) - removedSize)
    },
    pruned: true
  }
}
