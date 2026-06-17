import { describe, it, expect } from 'vitest'
import { galgameImageSrc } from './galgameImageSrc'

describe('galgameImageSrc', () => {
  it('prefers the server-injected cdn_url when present', () => {
    expect(
      galgameImageSrc({
        cdn_url: 'https://image.example/ab/cd/hash.webp',
        image_hash: 'abcd'
      })
    ).toBe('https://image.example/ab/cd/hash.webp')
  })

  it('falls back to the /image/<hash> token when cdn_url is missing', () => {
    // The "看不到图片" fix: a re-edit row hydrated WITHOUT cdn_url still
    // resolves from image_hash alone via the web SSR 302 middleware.
    expect(galgameImageSrc({ image_hash: 'deadbeef' })).toBe('/image/deadbeef')
  })

  it('falls back when cdn_url is an empty string', () => {
    expect(galgameImageSrc({ cdn_url: '', image_hash: 'feedface' })).toBe(
      '/image/feedface'
    )
  })
})
