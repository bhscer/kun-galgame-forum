// @vitest-environment nuxt
//
// galgameStatus drives status-badge labelling + wire-name resolution
// across Mine.vue, wizard search, PR review, wiki message center, etc.
// Nuxt env is needed because galgameNameFromWire reaches for the
// auto-imported getPreferredLanguageText (which itself consults the
// settings store for the user's locale priority).
import { describe, it, expect } from 'vitest'
import { galgameStatusBadge, galgameNameFromWire } from './galgameStatus'

describe('galgameStatusBadge', () => {
  it('maps each known status to the right (label,color)', () => {
    // GalgameStatus enum: 0=Published, 1=Banned, 2=VndbDraft, 3=Pending, 4=Declined
    expect(galgameStatusBadge(0)).toEqual({ label: '已发布', color: 'success' })
    expect(galgameStatusBadge(1)).toEqual({ label: '已封禁', color: 'default' })
    expect(galgameStatusBadge(2)).toEqual({
      label: 'VNDB 草稿',
      color: 'primary'
    })
    expect(galgameStatusBadge(3)).toEqual({ label: '待审核', color: 'warning' })
    expect(galgameStatusBadge(4)).toEqual({ label: '已拒绝', color: 'danger' })
  })

  it('returns 未知/default for unknown / undefined status (intentional)', () => {
    expect(galgameStatusBadge(undefined)).toEqual({
      label: '未知',
      color: 'default'
    })
    expect(galgameStatusBadge(99)).toEqual({ label: '未知', color: 'default' })
    // -1 is also unknown; never silently treated as Published.
    expect(galgameStatusBadge(-1)).toEqual({ label: '未知', color: 'default' })
  })
})

describe('galgameNameFromWire', () => {
  it('returns a localised name when at least one field is set', () => {
    const out = galgameNameFromWire({
      name_zh_cn: '标题',
      name_ja_jp: 'タイトル'
    })
    // The exact pick depends on locale priority; either is acceptable
    // as long as we don't return the fallback.
    expect(['标题', 'タイトル']).toContain(out)
  })

  it('returns fallback when all four name fields are missing/empty', () => {
    expect(galgameNameFromWire({}, '(无标题)')).toBe('(无标题)')
    expect(
      galgameNameFromWire(
        {
          name_en_us: '',
          name_ja_jp: '',
          name_zh_cn: '',
          name_zh_tw: ''
        },
        '#1'
      )
    ).toBe('#1')
  })

  it("default fallback is ''", () => {
    expect(galgameNameFromWire({})).toBe('')
  })

  it('snake_case wire keys are normalised to hyphenated locale keys', () => {
    // A subtle defect class: if name_en_us were forwarded without the
    // hyphenation, getPreferredLanguageText would never see it. Sanity-
    // check at least one path produces non-empty output for each
    // single-language input.
    expect(galgameNameFromWire({ name_en_us: 'Title' })).toBeTruthy()
    expect(galgameNameFromWire({ name_ja_jp: 'タイトル' })).toBeTruthy()
    expect(galgameNameFromWire({ name_zh_cn: '简中' })).toBeTruthy()
    expect(galgameNameFromWire({ name_zh_tw: '繁中' })).toBeTruthy()
  })
})
