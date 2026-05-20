// Parameterised revision-history hook. Drives the GalgameRevisionList
// component for ALL five entity types (galgame + 4 taxonomy entities)
// so the UI logic lives in one place; entity-specific routing is the
// only differentiator and it folds into a tiny lookup table.
//
// Endpoints (kungal proxy paths — kungal's path_mapper handles the
// /galgame/ vs bare /tag namespace difference; see path_mapper.go):
//
//   galgame: GET /galgame/:gid/history/all    → list (kungal-mapped to camelCase)
//            GET /galgame/:gid/revisions/:rev/diff  → diff
//            POST /galgame/:gid/revert        → revert
//
//   taxonomy: GET /galgame-<entity>/:id/revisions[/:rev] → list / single
//             POST /galgame-<entity>/:id/revert
//
// Note the asymmetry: galgame already exposed a kungal-mapped list
// endpoint (history/all → revisions, camelCase via GetGalgameHistory)
// AND a server-side diff endpoint (/revisions/:rev/diff). The taxonomy
// path is ProxyGet for list/snapshot — there's no diff endpoint on
// wiki for taxonomies in U3, so the FE computes the diff client-side
// by pulling two adjacent snapshots (or current vs picked) and feeding
// them to GalgameSnapshotDiff. List/snapshot routes are camelCase for
// galgame (kungal handler) but snake_case verbatim for taxonomies
// (ProxyGet) — `Revision` is normalised to a shared shape for the UI.

export type RevisionEntity =
  | 'galgame'
  | 'tag'
  | 'official'
  | 'engine'
  | 'series'

export interface GalgameRevisionListItem {
  id: number
  revision: number
  action: string
  note: string
  user: KunUser
  isMinor: boolean
  created: Date | string
}

// Internal: raw wiki taxonomy revision row (snake_case ProxyGet
// passthrough). We normalise to GalgameRevisionListItem at the
// boundary so the UI sees one shape.
interface WikiTaxonomyRevisionRow {
  id: number
  revision: number
  action: string
  note?: string
  user_id?: number
  user_role?: number
  changed_fields?: string[]
  ref_count?: number
  created: string
}

interface EndpointSet {
  list: string
  rev: (n: number) => string
  diff: ((n: number) => string) | null // null = no server-side diff
  revert: string
}

const endpointsFor = (entity: RevisionEntity, id: number): EndpointSet => {
  if (entity === 'galgame') {
    return {
      list: `/galgame/${id}/history/all`,
      rev: (n) => `/galgame/${id}/revisions/${n}`,
      diff: (n) => `/galgame/${id}/revisions/${n}/diff`,
      revert: `/galgame/${id}/revert`
    }
  }
  // taxonomy: kungal's suffix-aware path_mapper takes care of the
  // /galgame/<entity>/ namespace rewrite.
  return {
    list: `/galgame-${entity}/${id}/revisions`,
    rev: (n) => `/galgame-${entity}/${id}/revisions/${n}`,
    diff: null,
    revert: `/galgame-${entity}/${id}/revert`
  }
}

export const useRevisionHistory = (
  entity: RevisionEntity,
  id: Ref<number> | ComputedRef<number> | number
) => {
  const idRef = computed(() => (typeof id === 'number' ? id : unref(id)))
  const ep = computed(() => endpointsFor(entity, idRef.value))

  const pageData = reactive({ page: 1, limit: 10 })

  // List response shape differs by entity (galgame: kungal-mapped
  // {items, total} camelCase; taxonomy: wiki verbatim — could be a
  // bare array or {items,total} depending on wiki). Try the kungal
  // {items,total} shape first; the data ref unwraps `items` when
  // present, falls back to the bare array otherwise.
  type RawListResponse =
    | { items: GalgameRevisionListItem[]; total: number }
    | WikiTaxonomyRevisionRow[]
    | null

  const { data, status, refresh } = useKunFetch<RawListResponse>(
    () => ep.value.list,
    {
      lazy: true,
      method: 'GET',
      query: pageData,
      watch: [() => ep.value.list, () => pageData.page]
    }
  )

  // Normalise: taxonomy rows arrive snake_case + no embedded user
  // object; the UI only needs revision / action / note / created /
  // isMinor / user.name|avatar. We fabricate a stub user from
  // user_id when no embedded user is present (UI shows just the id);
  // a later wiki upgrade can embed users and this code will pick it
  // up automatically via the GalgameRevisionListItem cast.
  const items = computed<GalgameRevisionListItem[]>(() => {
    const d = data.value
    if (!d) return []
    if (Array.isArray(d)) {
      return d.map((r) => ({
        id: r.id,
        revision: r.revision,
        action: r.action,
        note: r.note ?? '',
        isMinor: false,
        created: r.created,
        user: {
          id: r.user_id ?? 0,
          name: r.user_id ? `#${r.user_id}` : '',
          avatar: ''
        } as KunUser
      }))
    }
    return d.items ?? []
  })

  const total = computed(() => {
    const d = data.value
    if (!d) return 0
    if (Array.isArray(d)) return d.length
    return d.total ?? 0
  })

  // Diff: galgame has a server-side endpoint; taxonomy doesn't, so we
  // construct one from two snapshots (target vs current — which the
  // UI obtains by fetching the LATEST revision separately).
  interface SnapshotDiff {
    changed_keys: Record<string, boolean>
    old: Record<string, unknown>
    new: Record<string, unknown>
  }

  const diffCache = reactive<Record<number, SnapshotDiff>>({})
  const diffLoading = ref<number | null>(null)

  const loadDiff = async (
    rev: number
  ): Promise<SnapshotDiff | null> => {
    if (diffCache[rev]) return diffCache[rev]!
    diffLoading.value = rev
    try {
      if (ep.value.diff) {
        // galgame: server-side diff endpoint returns the canonical shape.
        const res = await kunFetch<SnapshotDiff>(ep.value.diff(rev), {
          method: 'GET'
        })
        if (res) diffCache[rev] = res
        return res ?? null
      }
      // taxonomy: fetch the picked revision + current state (= latest
      // revision in the list — assumes list page 1 has it). Build a
      // diff by comparing the full snapshots; changed_keys = every
      // key whose JSON representation differs.
      const target = await kunFetch<{ snapshot: Record<string, unknown> }>(
        ep.value.rev(rev),
        { method: 'GET' }
      )
      if (!target) return null
      const latestRev = items.value[0]?.revision ?? rev
      const latest =
        latestRev === rev
          ? target
          : await kunFetch<{ snapshot: Record<string, unknown> }>(
              ep.value.rev(latestRev),
              { method: 'GET' }
            )
      if (!latest) return null
      const oldSnap = target.snapshot ?? {}
      const newSnap = latest.snapshot ?? {}
      const allKeys = new Set([
        ...Object.keys(oldSnap),
        ...Object.keys(newSnap)
      ])
      const changed_keys: Record<string, boolean> = {}
      for (const k of allKeys) {
        if (JSON.stringify(oldSnap[k]) !== JSON.stringify(newSnap[k])) {
          changed_keys[k] = true
        }
      }
      const diff = { changed_keys, old: oldSnap, new: newSnap }
      diffCache[rev] = diff
      return diff
    } finally {
      diffLoading.value = null
    }
  }

  const reverting = ref<number | null>(null)

  const revert = async (rev: number): Promise<boolean> => {
    reverting.value = rev
    try {
      const res = await kunFetch(ep.value.revert, {
        method: 'POST',
        body: { revision: rev }
      })
      if (res !== null) {
        // Bust the local diff cache (latest may have shifted) and
        // refresh the list so the new 'reverted' revision shows up.
        // dynamic-delete is unavoidable here — diffCache is a reactive
        // proxy and reassigning would lose its identity; clearing keys
        // one by one is the supported way.
        for (const k of Object.keys(diffCache)) {
          // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
          delete diffCache[Number(k)]
        }
        await refresh()
        return true
      }
      return false
    } finally {
      reverting.value = null
    }
  }

  return {
    items,
    total,
    status,
    pageData,
    refresh,
    loadDiff,
    diffCache,
    diffLoading,
    revert,
    reverting
  }
}
