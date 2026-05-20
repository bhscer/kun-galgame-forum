import { describe, it, expect } from 'vitest'
import { arrayDiffByKey, arrayDiffScalar } from './arrayDiffByKey'

describe('arrayDiffByKey', () => {
  const k = (x: { id: number }) => String(x.id)

  it('pure add / remove / changed split', () => {
    const oldArr = [
      { id: 1, label: 'a', sort: 0 },
      { id: 2, label: 'b', sort: 1 }
    ]
    const newArr = [
      { id: 1, label: 'a', sort: 0 }, // unchanged
      { id: 3, label: 'c', sort: 2 } // added; 2 removed
    ]
    const r = arrayDiffByKey(oldArr, newArr, k)
    expect(r.added.map((x) => x.id)).toEqual([3])
    expect(r.removed.map((x) => x.id)).toEqual([2])
    expect(r.changed).toHaveLength(0)
  })

  it('detects scalar field changes on matched key', () => {
    const oldArr = [{ id: 1, label: 'a', sort: 0 }]
    const newArr = [{ id: 1, label: 'A', sort: 5 }]
    const r = arrayDiffByKey(oldArr, newArr, k)
    expect(r.added).toHaveLength(0)
    expect(r.removed).toHaveLength(0)
    expect(r.changed).toHaveLength(1)
    expect(r.changed[0]!.fields.sort()).toEqual(['label', 'sort'])
  })

  it('compareFields restricts the change check', () => {
    const oldArr = [{ id: 1, label: 'a', sort: 0 }]
    const newArr = [{ id: 1, label: 'A', sort: 5 }] // both differ
    const r = arrayDiffByKey(oldArr, newArr, k, ['sort']) // ignore label
    expect(r.changed[0]!.fields).toEqual(['sort'])
  })

  it('handles nil / empty arrays without throwing', () => {
    expect(arrayDiffByKey(null, null, k)).toEqual({
      added: [],
      removed: [],
      changed: []
    })
    const r = arrayDiffByKey([], [{ id: 1, label: 'x', sort: 0 }], k)
    expect(r.added).toHaveLength(1)
  })

  it('deep-equals nested objects on matched key (no false positive)', () => {
    const oldArr = [{ id: 1, meta: { x: 1, y: [1, 2] } }]
    const newArr = [{ id: 1, meta: { x: 1, y: [1, 2] } }]
    const r = arrayDiffByKey(oldArr, newArr, k)
    expect(r.changed).toHaveLength(0)
  })

  it('detects nested object differences via JSON fallback', () => {
    const oldArr = [{ id: 1, meta: { x: 1 } }]
    const newArr = [{ id: 1, meta: { x: 2 } }]
    const r = arrayDiffByKey(oldArr, newArr, k)
    expect(r.changed[0]!.fields).toEqual(['meta'])
  })
})

describe('arrayDiffScalar', () => {
  it('symmetric difference for numbers', () => {
    const r = arrayDiffScalar([1, 2, 3], [2, 3, 4])
    expect(r.added).toEqual([4])
    expect(r.removed).toEqual([1])
  })

  it('symmetric difference for strings, order-independent + dedup', () => {
    const r = arrayDiffScalar(['a', 'b', 'b'], ['b', 'c'])
    expect(r.added).toEqual(['c'])
    expect(r.removed).toEqual(['a'])
  })

  it('handles nil + empty', () => {
    expect(arrayDiffScalar(null, undefined)).toEqual({
      added: [],
      removed: []
    })
    expect(arrayDiffScalar([], [1, 2])).toEqual({ added: [1, 2], removed: [] })
  })
})
