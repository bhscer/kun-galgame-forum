export interface GalgameStorePersist {
  vndbId: string
  name: KunLanguage
  introduction: KunLanguage
  contentLimit: 'sfw' | 'nsfw'
  ageLimit: 'all' | 'r18'
  originalLanguage: Language
  aliases: string[]
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
}
