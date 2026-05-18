// PR list item (kungal maps /galgame/:gid/pr/all → camelCase).
export interface GalgamePR {
  id: number
  galgameId: number
  status: number
  note: string
  baseRevision: number
  user: KunUser
  completedTime: Date | string | null
  created: Date | string
}

// Raw wiki PR detail response: GET /galgame/:gid/prs/:id (ProxyGet,
// passed through verbatim — snake_case). See docs 02-revisions-and-prs.
export interface WikiPRDetailResponse {
  pr: {
    id: number
    galgame_id: number
    user_id: number
    status: number
    note: string
    base_revision: number
    snapshot: Record<string, unknown>
    completed_by: number | null
    revision_id: number | null
    created: string
  }
  changed_keys: Record<string, boolean>
}

// Normalized shape Info.vue builds and Details.vue renders: old =
// base-revision snapshot, new = pr.snapshot, limited to changed_keys.
export interface GalgamePRDiffView {
  id: number
  galgameId: number
  status: number
  note: string
  changedKeys: Record<string, boolean>
  oldSnap: Record<string, unknown>
  newSnap: Record<string, unknown>
}
