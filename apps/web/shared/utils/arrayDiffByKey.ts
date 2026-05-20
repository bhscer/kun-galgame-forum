// arrayDiffByKey — structural diff for arrays of objects keyed by a
// stable identity. Drives SnapshotDiff's rendering of presence-replace
// relations (covers/screenshots/links): instead of stringifying both
// sides and running an LCS char-diff (which produces unreadable JSON
// blobs for any non-trivial change), we report per-item:
//
//   - added:   rows present in new but not in old
//   - removed: rows present in old but not in new
//   - changed: same key, different scalar fields → list of fields that
//              differ (so the renderer can show field-level deltas)
//
// "Same key" is determined by the caller-provided keyFn — e.g. covers
// use image_hash, links use `${name}|||${link}`. compareFields lets the
// caller restrict the "changed" check to relevant scalars (e.g. ignore
// the server-injected cdn_url derivation field).

export interface ArrayChangedItem<T> {
  key: string
  old: T
  new: T
  fields: string[]
}

export interface ArrayDiffResult<T> {
  added: T[]
  removed: T[]
  changed: ArrayChangedItem<T>[]
}

export const arrayDiffByKey = <T>(
  oldArr: readonly T[] | undefined | null,
  newArr: readonly T[] | undefined | null,
  keyFn: (item: T) => string,
  compareFields?: (keyof T)[]
): ArrayDiffResult<T> => {
  const oldMap = new Map<string, T>()
  for (const item of oldArr ?? []) oldMap.set(keyFn(item), item)

  const newMap = new Map<string, T>()
  for (const item of newArr ?? []) newMap.set(keyFn(item), item)

  const added: T[] = []
  const removed: T[] = []
  const changed: ArrayChangedItem<T>[] = []

  for (const [k, n] of newMap) {
    const o = oldMap.get(k)
    if (!o) {
      added.push(n)
      continue
    }
    // Same key — compare scalar fields. If compareFields is given,
    // only those are inspected; otherwise iterate all top-level keys
    // of the new row. Object/array nested values fall back to JSON
    // equality (good enough for snapshot diffs since each row is small).
    const fieldsToCheck =
      compareFields ??
      (Object.keys(n as Record<string, unknown>) as (keyof T)[])
    const diffFields: string[] = []
    for (const f of fieldsToCheck) {
      const ov = (o as Record<string, unknown>)[f as string]
      const nv = (n as Record<string, unknown>)[f as string]
      const same =
        typeof ov === 'object' || typeof nv === 'object'
          ? JSON.stringify(ov) === JSON.stringify(nv)
          : ov === nv
      if (!same) diffFields.push(String(f))
    }
    if (diffFields.length) {
      changed.push({ key: k, old: o, new: n, fields: diffFields })
    }
  }

  for (const [k, o] of oldMap) {
    if (!newMap.has(k)) removed.push(o)
  }

  return { added, removed, changed }
}

// arrayDiffScalar — set-difference for arrays of primitive values
// (tag_ids / official_ids / engine_ids / aliases). Order doesn't
// matter; duplicates collapse. Returns the symmetric difference split
// into added/removed.
export const arrayDiffScalar = <T extends string | number>(
  oldArr: readonly T[] | undefined | null,
  newArr: readonly T[] | undefined | null
): { added: T[]; removed: T[] } => {
  const oldSet = new Set<T>(oldArr ?? [])
  const newSet = new Set<T>(newArr ?? [])
  const added: T[] = []
  const removed: T[] = []
  for (const v of newSet) if (!oldSet.has(v)) added.push(v)
  for (const v of oldSet) if (!newSet.has(v)) removed.push(v)
  return { added, removed }
}
