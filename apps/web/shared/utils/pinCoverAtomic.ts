// Atomic "pin new cover as banner" transformation.
//
// Wiki has a partial unique index `idx_galgame_cover_pinned` enforcing
// at-most-one row with `sort_order = 0` per galgame. Naive two-step
// "set new=0, demote old" through two separate PUTs would either be
// rejected mid-flight (both rows momentarily 0) or leave a stale 0-row
// if the second step fails. To keep the whole edit a single atomic
// PUT, we compute the new covers[] locally:
//
//   - target row → sort_order = 0
//   - whatever row was previously sort_order=0 → sort_order = max+1
//   - other rows unchanged
//
// The whole array is submitted as a presence-replace, so wiki processes
// the demotion + promotion in one transaction and the partial unique
// index is never violated.
//
// Idempotency: pinning the already-pinned row is a no-op. Pinning when
// no current pinned row exists (corner case: orphaned set) just sets
// target to 0 without demoting anything.

interface CoverLike {
  image_hash: string
  sort_order: number
}

export const pinCoverAtomic = <T extends CoverLike>(
  covers: readonly T[],
  targetHash: string
): T[] => {
  const target = covers.find((c) => c.image_hash === targetHash)
  if (!target || target.sort_order === 0) {
    // Either the hash isn't in the set, or it's already pinned —
    // both cases are no-ops (return a fresh shallow copy so callers
    // can use it as a v-model write without aliasing concerns).
    return covers.slice()
  }
  // Highest sort_order to demote the previous pinned row past.
  // Start from 0 — if there are no other rows, the demoted previous
  // pinned still ends up at 1 (max+1 with max=0 from itself).
  const maxOrder = covers.reduce(
    (m, c) => (c.sort_order > m ? c.sort_order : m),
    0
  )
  return covers.map((c) => {
    if (c.image_hash === targetHash) return { ...c, sort_order: 0 }
    if (c.sort_order === 0) return { ...c, sort_order: maxOrder + 1 }
    return c
  })
}
