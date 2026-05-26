// Single source of truth for "which URL to use as a galgame's head image".
//
// Kungal's rewriteBanners walker resolves `effective_banner_hash`
// (covers[sort_order=0]) into `effective_banner_url` on the server
// side (see apps/api/internal/galgame/client/banner.go). The FE picks:
//
//   1. `effective_banner_url` — the U2 canonical head image. Preferred
//      whenever present.
//   2. `banner` — legacy free-form URL on old (pre-U2) galgame rows
//      that haven't been migrated to covers. Kept as fallback for
//      historical data — wiki still emits the field.
//
// Returning `''` means "no banner available" — callers can show a
// placeholder. We intentionally do NOT fall back to a hardcoded default
// URL: that's the caller's policy decision (some surfaces want a CDN
// placeholder, others want a Nuxt /placeholder.webp asset).
//
// (banner_image_hash was retired in wiki PR5 / K-PR6 — no longer a
// fallback layer.)
interface BannerSource {
  effective_banner_url?: string
  banner?: string
}

// Two CDN conventions coexist during the legacy → image_service
// migration of galgame covers:
//
//   image_service (new, hash-addressed):
//     main    {cdn}/{hash[:2]}/{hash[2:4]}/{hash}.webp           1920×1080
//     variant {cdn}/{hash[:2]}/{hash[2:4]}/{hash}_{variant}.webp  underscore
//
//   legacy nitro (pre-migration, path-based, e.g. image.kungal.com):
//     main    {cdn}/galgame/{id}/banner/banner.webp
//     variant {cdn}/galgame/{id}/banner/banner-{variant}.webp     hyphen
//
// Old galgame rows still serve from the legacy host and won't move
// until the bulk migration runs, so the helper must produce the right
// URL for either. We detect which family a URL belongs to by its path
// shape — image_service URLs always end with a hex-hash file under a
// two-level hex prefix; anything else (including arbitrary external
// URLs) is treated as legacy / unknown and uses the hyphen form.
//
// Non-`.webp` URLs are returned untouched — historical free-form
// `banner` fallbacks may be jpg/png from arbitrary external sources
// where no variant convention applies; rewriting them would 404.
// All variants currently in use across the app:
//   'mini'  — galgame banner mini (460×259)
//   '100'   — avatar small (100×100, comment lists, top bar)
//   '256'   — avatar medium (256×256, profile cards)
type ImageVariant = 'mini' | '100' | '256'

const IMAGE_SERVICE_HASH_PATH = /\/[0-9a-f]{2}\/[0-9a-f]{2}\/[0-9a-f]+\.webp$/i

// Exposed so call sites that receive a raw URL string (banner from
// series detail, avatar from OAuth callback / store, etc.) can reuse
// the same convention detection without re-implementing it. Detects
// image_service URLs by their hash-addressed path shape and applies
// underscore; everything else (legacy nitro paths, arbitrary external
// URLs) gets the hyphen form.
export const withImageVariant = (
  url: string,
  variant: ImageVariant
): string => {
  if (!url || !/\.webp$/i.test(url)) return url
  const sep = IMAGE_SERVICE_HASH_PATH.test(url) ? '_' : '-'
  return url.replace(/\.webp$/i, `${sep}${variant}.webp`)
}

// Back-compat alias — older call sites import `withBannerVariant` and
// pass the banner variant union. Keep it as a thin wrapper so existing
// imports keep compiling; internal logic is shared with the generic.
export const withBannerVariant = (
  url: string,
  variant: Extract<ImageVariant, 'mini'>
): string => withImageVariant(url, variant)

export const getEffectiveBanner = (
  g?: BannerSource | null,
  opts?: { variant?: Extract<ImageVariant, 'mini'> }
): string => {
  if (!g) return ''
  const eff = g.effective_banner_url?.trim()
  const base = (eff || g.banner?.trim() || '').trim()
  if (!base || !opts?.variant) return base
  return withBannerVariant(base, opts.variant)
}
