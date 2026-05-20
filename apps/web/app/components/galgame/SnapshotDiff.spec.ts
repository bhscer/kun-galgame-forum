// @vitest-environment nuxt
//
// SnapshotDiff structural-diff component spec. The Nuxt environment is
// needed because the component reaches for the auto-imported KunBadge /
// KunNull / useDiff symbols at render time; happy-dom alone would
// resolve them to undefined and the test would assert against an empty
// HTML shell.
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import SnapshotDiff from './SnapshotDiff.vue'

describe('SnapshotDiff', () => {
  it('renders KunNull when changedKeys is empty', async () => {
    const w = await mountSuspended(SnapshotDiff, {
      props: { changedKeys: {}, oldSnap: {}, newSnap: {} }
    })
    expect(w.html()).toContain('无字段变化')
  })

  it('scalar diff: emits both old and new fragments', async () => {
    const w = await mountSuspended(SnapshotDiff, {
      props: {
        changedKeys: { name_zh_cn: true },
        oldSnap: { name_zh_cn: '旧标题' },
        newSnap: { name_zh_cn: '新标题' }
      }
    })
    const html = w.html()
    // useDiff marks added as <b> and removed as <strong>; the label
    // should also surface from the field map.
    expect(html).toContain('简体中文标题')
    expect(html).toMatch(/<b>|<strong>/)
  })

  it('array-of-scalars: tag_ids renders ➕/➖ badges, not raw JSON', async () => {
    const w = await mountSuspended(SnapshotDiff, {
      props: {
        changedKeys: { tag_ids: true },
        oldSnap: { tag_ids: [1, 2, 3] },
        newSnap: { tag_ids: [2, 3, 4] }
      }
    })
    const html = w.html()
    expect(html).toContain('➕')
    expect(html).toContain('4')
    expect(html).toContain('➖')
    expect(html).toContain('1')
    // Critically: must NOT fall back to JSON-LCS garble.
    expect(html).not.toContain('[1,2,3]')
  })

  it('array-of-objects (covers): added / removed / changed split', async () => {
    const w = await mountSuspended(SnapshotDiff, {
      props: {
        changedKeys: { covers: true },
        oldSnap: {
          covers: [
            { image_hash: 'aaaa', sort_order: 0, sexual: 0, violence: 0 },
            { image_hash: 'bbbb', sort_order: 1, sexual: 0, violence: 0 }
          ]
        },
        newSnap: {
          covers: [
            // aaaa changed sort_order
            { image_hash: 'aaaa', sort_order: 5, sexual: 0, violence: 0 },
            // bbbb removed; cccc added
            { image_hash: 'cccc', sort_order: 1, sexual: 0, violence: 0 }
          ]
        }
      }
    })
    const html = w.html()
    expect(html).toContain('新增')
    expect(html).toContain('cccc')
    expect(html).toContain('删除')
    expect(html).toContain('bbbb')
    expect(html).toContain('修改')
    expect(html).toContain('sort_order')
  })

  it('array-of-objects (links): composite key — same (name,link) not a change', async () => {
    const w = await mountSuspended(SnapshotDiff, {
      props: {
        changedKeys: { links: true },
        oldSnap: {
          links: [
            { name: 'a', link: 'x' },
            { name: 'b', link: 'y' }
          ]
        },
        newSnap: {
          links: [
            { name: 'b', link: 'y' }, // reorder only
            { name: 'a', link: 'x' }
          ]
        }
      }
    })
    // Pure reorder produces no add/remove/changed → row hidden →
    // KunNull placeholder kicks in.
    expect(w.html()).toContain('无字段变化')
  })

  it('falls back gracefully on an unknown array-of-scalars field', async () => {
    const w = await mountSuspended(SnapshotDiff, {
      props: {
        changedKeys: { some_future_ids: true },
        oldSnap: { some_future_ids: [10, 20] },
        newSnap: { some_future_ids: [20, 30] }
      }
    })
    expect(w.html()).toContain('➕')
    expect(w.html()).toContain('30')
  })
})
