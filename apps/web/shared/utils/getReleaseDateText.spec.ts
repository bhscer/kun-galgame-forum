import { describe, it, expect } from 'vitest'
import { getReleaseDateText } from './getReleaseDateText'

describe('getReleaseDateText', () => {
  it('未公布 when no date and not TBA', () => {
    expect(getReleaseDateText(null, false)).toBe('未公布')
    expect(getReleaseDateText('', false)).toBe('未公布')
    expect(getReleaseDateText(undefined, undefined)).toBe('未公布')
  })

  it('未定 (TBA) when no date but TBA flagged', () => {
    expect(getReleaseDateText(null, true)).toBe('未定 (TBA)')
    expect(getReleaseDateText('', true)).toBe('未定 (TBA)')
  })

  it('returns the date when present and not TBA', () => {
    expect(getReleaseDateText('2024-06-15', false)).toBe('2024-06-15')
    expect(getReleaseDateText('2024-06-15')).toBe('2024-06-15')
  })

  it('predicted prefix when date present + TBA flagged', () => {
    expect(getReleaseDateText('2024-06', true)).toBe('预计 2024-06')
    expect(getReleaseDateText('2024-06-15', true)).toBe('预计 2024-06-15')
  })

  it('trims whitespace-only date strings', () => {
    expect(getReleaseDateText('   ', false)).toBe('未公布')
    expect(getReleaseDateText('   ', true)).toBe('未定 (TBA)')
  })
})
