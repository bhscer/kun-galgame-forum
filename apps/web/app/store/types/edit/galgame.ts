export interface GalgameStorePersist {
  vndbId: string
  name: KunLanguage
  introduction: KunLanguage
  contentLimit: 'sfw' | 'nsfw'
  ageLimit: 'all' | 'r18'
  originalLanguage: Language
  aliases: string[]
  // U1: "" = unknown; sent as wire `release_date` (snake_case at boundary
  // by Footer). releaseDateTBA can coexist with a concrete date (semantic
  // "predicted Y/M").
  releaseDate: string
  releaseDateTBA: boolean
}

// PR-edit working copy (also reused by Draft.vue). Backs the temp,
// non-persisted store: it is fully re-hydrated from the galgame detail
// every time the user opens the rewrite form, so leftover state never
// bleeds across edits.
//
// tags/officials/engines hold the FULL entity objects (not just ids) so
// the selectors can render names without a second lookup; Footer.vue
// derives the *_ids[] wire arrays at submit time. links mirrors the wiki
// PR `links` shape ({name, link}); note is the PR description.
//
// IMPORTANT (wiki "replace-all" semantics): aliases / tag_ids /
// official_ids / engine_ids / links each REPLACE the entire set on
// merge. These must be pre-hydrated COMPLETELY from current data —
// see components/galgame/Rewrite.vue — or a submit silently wipes them.
export interface GalgameEditStoreTemp {
  id: number
  vndbId: string
  name: KunLanguage
  introduction: KunLanguage
  contentLimit: 'sfw' | 'nsfw'
  ageLimit: 'all' | 'r18'
  originalLanguage: Language
  alias: string[]
  tags: GalgameTagItem[]
  officials: GalgameOfficialItem[]
  engines: GalgameEngineItem[]
  links: { name: string; link: string }[]
  note: string
  // U1: "" = unknown (cleared); concrete "YYYY-MM-DD" or any (TBA may set
  // a predicted partial date — wiki accepts the same string format).
  releaseDate: string
  releaseDateTBA: boolean
  // U2: cover candidate set + screenshot gallery, presence-replace
  // (wiki PUT/PR treats the whole array as authoritative). Pre-hydrated
  // COMPLETELY from galgame.covers/screenshots in Rewrite.vue — drop a
  // row by removing from the array, add new by uploading then pushing.
  // cdn_url is a server-injected derived field for previews; Footer
  // strips it before sending to wiki (wiki doesn't accept it on write).
  covers: GalgameCover[]
  screenshots: GalgameScreenshot[]
  // True when the current user is the galgame's creator or an
  // admin/moderator (role>=2): wiki lets them edit directly via
  // PUT /galgame/:gid (instant, new revision) instead of opening a PR.
  // Computed once at hydration (Rewrite.vue has galgame.user + the user
  // store); Footer.vue branches the submit endpoint on it. Draft.vue
  // sets it false (drafts always PATCH, never this path).
  canDirectEdit: boolean
}
