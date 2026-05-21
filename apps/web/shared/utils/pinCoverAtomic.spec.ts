import { describe, it, expect } from 'vitest'
import { pinCoverAtomic } from './pinCoverAtomic'

interface Cover {
  image_hash: string
  sort_order: number
  extra?: string
}

describe('pinCoverAtomic', () => {
  it('promotes target → 0 and demotes the previous pinned past max', () => {
    const covers: Cover[] = [
      { image_hash: 'old', sort_order: 0 },
      { image_hash: 'mid', sort_order: 1 },
      { image_hash: 'tgt', sort_order: 2 }
    ]
    const out = pinCoverAtomic(covers, 'tgt')
    const byHash = Object.fromEntries(out.map((c) => [c.image_hash, c.sort_order]))
    expect(byHash['tgt']).toBe(0)
    expect(byHash['old']).toBe(3) // max+1 = 2+1
    expect(byHash['mid']).toBe(1) // untouched
  })

  it('is a no-op when target is already pinned', () => {
    const covers: Cover[] = [
      { image_hash: 'tgt', sort_order: 0 },
      { image_hash: 'b', sort_order: 1 }
    ]
    const out = pinCoverAtomic(covers, 'tgt')
    expect(out).toEqual(covers)
    // Returns a fresh copy (different array identity) — safe for v-model.
    expect(out).not.toBe(covers)
  })

  it('is a no-op when target hash not in set', () => {
    const covers: Cover[] = [{ image_hash: 'a', sort_order: 0 }]
    const out = pinCoverAtomic(covers, 'missing')
    expect(out).toEqual(covers)
  })

  it('sets target=0 with no demote when no current 0-row exists', () => {
    // Corner case: orphan set (no pinned cover). Pinning target should
    // just set it to 0 without touching others.
    const covers: Cover[] = [
      { image_hash: 'a', sort_order: 1 },
      { image_hash: 'b', sort_order: 2 }
    ]
    const out = pinCoverAtomic(covers, 'b')
    const byHash = Object.fromEntries(out.map((c) => [c.image_hash, c.sort_order]))
    expect(byHash['b']).toBe(0)
    expect(byHash['a']).toBe(1) // untouched
  })

  it('preserves extra fields via spread (no data loss)', () => {
    const covers: Cover[] = [
      { image_hash: 'a', sort_order: 0, extra: 'keep-a' },
      { image_hash: 'b', sort_order: 1, extra: 'keep-b' }
    ]
    const out = pinCoverAtomic(covers, 'b')
    expect(out.find((c) => c.image_hash === 'a')?.extra).toBe('keep-a')
    expect(out.find((c) => c.image_hash === 'b')?.extra).toBe('keep-b')
  })

  it('partial unique index invariant: never produces two rows at sort_order=0', () => {
    const covers: Cover[] = [
      { image_hash: 'old', sort_order: 0 },
      { image_hash: 'tgt', sort_order: 5 }
    ]
    const out = pinCoverAtomic(covers, 'tgt')
    const zeros = out.filter((c) => c.sort_order === 0)
    expect(zeros).toHaveLength(1)
    expect(zeros[0]!.image_hash).toBe('tgt')
  })
})
