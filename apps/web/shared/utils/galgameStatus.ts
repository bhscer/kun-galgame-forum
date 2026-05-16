import { GalgameStatus } from '../types/galgame'
import { getPreferredLanguageText } from './getPreferredLanguageText'

// Centralised galgame submission-status presentation. Every consumer
// (publish wizard, my-submissions list, draft editor, admin review
// queue, wiki notifications) renders status the same way, and a new
// status value only needs touching here.
//
// Rationale: the "VNDB 草稿 shown with a 前往发布资源 button" bug came
// from each component carrying its own copy of this mapping, so one
// place (the wizard's search-items block) lagged behind the others.
// Keeping the status→{label,color} and wire→name logic in one module
// makes that class of drift impossible.

// Subset of KunUIColor actually used for status badges. Assignable to
// KunBadge's `color` prop (which is the full KunUIColor union).
type GalgameStatusColor =
  | 'default'
  | 'primary'
  | 'success'
  | 'warning'
  | 'danger'

export interface GalgameStatusBadge {
  label: string
  color: GalgameStatusColor
}

// status may be undefined for rows where the wire format omits it
// (older briefs / partial projections) — fall through to "未知" rather
// than guessing "已发布", which would mislabel hidden drafts.
export const galgameStatusBadge = (
  status: number | undefined
): GalgameStatusBadge => {
  switch (status) {
    case GalgameStatus.Published:
      return { label: '已发布', color: 'success' }
    case GalgameStatus.Banned:
      return { label: '已封禁', color: 'default' }
    case GalgameStatus.VndbDraft:
      return { label: 'VNDB 草稿', color: 'primary' }
    case GalgameStatus.Pending:
      return { label: '待审核', color: 'warning' }
    case GalgameStatus.Declined:
      return { label: '已拒绝', color: 'danger' }
    default:
      return { label: '未知', color: 'default' }
  }
}

// Wire-format galgame rows carry snake_case name_<locale> columns; the
// rest of the site keys KunLanguage by hyphenated locale. Build the
// hyphen shape and run it through the standard locale-priority picker
// so titles resolve identically to GalgameCard / detail pages.
export interface WireGalgameName {
  name_en_us?: string
  name_ja_jp?: string
  name_zh_cn?: string
  name_zh_tw?: string
}

export const galgameNameFromWire = (
  g: WireGalgameName,
  fallback = ''
): string => {
  return (
    getPreferredLanguageText({
      'en-us': g.name_en_us ?? '',
      'ja-jp': g.name_ja_jp ?? '',
      'zh-cn': g.name_zh_cn ?? '',
      'zh-tw': g.name_zh_tw ?? ''
    }) || fallback
  )
}
