// Resolve a cover/screenshot row to a displayable image URL.
//
// Prefer the server-injected `cdn_url` (kungal's rewriteBanners walker) when
// present — no redirect. Otherwise fall back to the domain-independent
// `/image/<hash>` token, which the web SSR middleware 302-redirects to the
// active image CDN (server/middleware/content-image.ts). This makes the editor
// render the image from `image_hash` ALONE, so a row that arrived WITHOUT
// `cdn_url` (e.g. a re-edit hydration that didn't pass through the walker)
// still shows the picture instead of an empty box — the "看不到图片" fix.
export const galgameImageSrc = (row: {
  cdn_url?: string
  image_hash: string
}): string => row.cdn_url || `/image/${row.image_hash}`
