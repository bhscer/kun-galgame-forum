// TODO: rebuild dynamic sitemap routes via the Go API.
//
// Original implementation (pre-Go-rewrite) loaded everything from a
// shared Prisma client — topics, galgames, taxonomy, websites, toolsets,
// ratings. Since the DB layer moved entirely to Go Fiber + GORM and the
// Nuxt frontend no longer has any direct DB access, this generator
// needs to be rewritten as a series of fetches against the public Go
// API endpoints (`/api/topic`, `/api/galgame`, `/api/website`, etc.)
// before sitemap output is restored.
//
// Returning an empty array keeps generateKunSitemap.ts working so the
// rest of the build pipeline doesn't break; the sitemap will only carry
// the static routes Nuxt knows about until this is rewritten.
export const getKunDynamicRoutes = async (): Promise<
  { path: string; lastmod: string }[]
> => {
  return []
}
