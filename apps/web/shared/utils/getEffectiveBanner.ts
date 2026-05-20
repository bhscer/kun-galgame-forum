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

export const getEffectiveBanner = (g?: BannerSource | null): string => {
  if (!g) return ''
  const eff = g.effective_banner_url?.trim()
  if (eff) return eff
  return g.banner?.trim() ?? ''
}
