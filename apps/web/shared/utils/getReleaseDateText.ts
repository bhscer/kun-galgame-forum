// Single source of truth for rendering a galgame's release date.
// Centralised so display points (detail page, cards, "我的提交" list,
// admin queues) never drift into bespoke fallback ladders.
//
// Wire contract (U1): { release_date: string | null, release_date_tba: bool }.
// - nil/empty + TBA=false → "未公布"
// - nil/empty + TBA=true  → "未定 (TBA)"
// - "YYYY-MM-DD" + TBA=false → "YYYY-MM-DD"
// - "YYYY-MM-DD" + TBA=true  → "预计 YYYY-MM-DD"  ← TBA may coexist
//                                                   with a predicted date
//                                                   (wiki design choice).
export const getReleaseDateText = (
  releaseDate?: string | null,
  releaseDateTBA?: boolean
): string => {
  const d = (releaseDate ?? '').trim()
  if (!d) return releaseDateTBA ? '未定 (TBA)' : '未公布'
  return releaseDateTBA ? `预计 ${d}` : d
}
