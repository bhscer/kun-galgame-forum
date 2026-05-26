// Build-time sitemap generator. Pre-Go-rewrite this loaded everything
// directly from a shared Prisma client; we now fetch the same data over
// the public Go API.
//
// SEO safety:
//   - All list fetches force the `KUNGalgameSettings` cookie to
//     `showKUNGalgameContentLimit=sfw` so the BE applies its SFW filter
//     (utils.IsSFW). This guarantees NSFW detail URLs never enter the
//     sitemap regardless of what's available in the DB. Search engines
//     don't carry cookies, so the live crawl path is also SFW; the
//     sitemap just makes that explicit.
//   - Failed endpoints fall back to "no rows for this kind" rather than
//     aborting the whole build — a partial sitemap is better than no
//     sitemap (Nuxt would still emit static routes).

const API = process.env.KUN_GALGAME_API
if (!API) {
  // Without the API URL we can't fetch anything; surface the misconfig
  // loudly rather than silently emitting an empty sitemap.
  throw new Error('KUN_GALGAME_API env var is required for sitemap build')
}

interface KunDynamicRoute {
  path: string
  lastmod: string
}

// Forces the BE to treat us as a SFW-cookie client. The exact JSON the
// Pinia store persists; matches utils.IsSFW(c) on the Go side.
const SFW_COOKIE = `KUNGalgameSettings=${encodeURIComponent(
  JSON.stringify({ showKUNGalgameContentLimit: 'sfw' })
)}`

const baseHeaders: HeadersInit = {
  Cookie: SFW_COOKIE,
  Accept: 'application/json'
}

const nowIso = () => new Date().toISOString()

const toIso = (d: unknown): string => {
  if (typeof d === 'string' || typeof d === 'number' || d instanceof Date) {
    const t = new Date(d)
    if (!Number.isNaN(t.getTime())) return t.toISOString()
  }
  return nowIso()
}

// Generic paginated fetch. Walks pages until the BE either runs out or
// we hit `maxRows` (safety cap to avoid an unbounded loop if a list
// endpoint mis-reports total).
const fetchPaginated = async <T>(
  path: string,
  pickItems: (json: unknown) => T[],
  pickTotal: (json: unknown) => number,
  pageSize = 50,
  maxRows = 5000
): Promise<T[]> => {
  const out: T[] = []
  for (let page = 1; out.length < maxRows; page++) {
    const url = new URL(API + path, 'http://x')
    url.searchParams.set('page', String(page))
    url.searchParams.set('limit', String(pageSize))
    let json: unknown
    try {
      const res = await fetch(url.pathname + url.search, {
        headers: baseHeaders
      })
      if (!res.ok) break
      json = await res.json()
    } catch {
      break
    }
    const items = pickItems(json)
    if (items.length === 0) break
    out.push(...items)
    const total = pickTotal(json)
    if (out.length >= total) break
  }
  return out.slice(0, maxRows)
}

// Most kungal endpoints wrap the payload in `{ code, message, data }`.
const unwrap = (json: unknown): unknown =>
  (json as { data?: unknown })?.data ?? json

// Topic list — /api/topic, paginated.
const fetchTopics = (): Promise<KunDynamicRoute[]> =>
  fetchPaginated<KunDynamicRoute>(
    '/topic',
    (json) => {
      const data = unwrap(json) as { topics?: unknown[] } | unknown[]
      const items = Array.isArray(data)
        ? (data as { id: number; statusUpdateTime?: string }[])
        : ((data?.topics ?? []) as {
            id: number
            statusUpdateTime?: string
          }[])
      return items.map((t) => ({
        path: `/topic/${t.id}`,
        lastmod: toIso(t.statusUpdateTime)
      }))
    },
    (json) => {
      const data = unwrap(json) as { total?: number } | unknown
      return (data as { total?: number })?.total ?? Number.MAX_SAFE_INTEGER
    }
  )

// Galgame list — /api/galgame, paginated. Service-layer SFW filter is
// in place (galgame_service.GetList), so SFW cookie alone is enough.
const fetchGalgames = (): Promise<KunDynamicRoute[]> =>
  fetchPaginated<KunDynamicRoute>(
    '/galgame',
    (json) => {
      const data = unwrap(json) as
        | { galgames?: { id: number; resourceUpdateTime?: string }[] }
        | undefined
      return (data?.galgames ?? []).map((g) => ({
        path: `/galgame/${g.id}`,
        lastmod: toIso(g.resourceUpdateTime)
      }))
    },
    (json) => (unwrap(json) as { total?: number })?.total ?? 0
  )

// Galgame resources — /api/galgame-resource (SFW-filtered service-side).
const fetchResources = (): Promise<KunDynamicRoute[]> =>
  fetchPaginated<KunDynamicRoute>(
    '/galgame-resource',
    (json) => {
      const data = unwrap(json) as
        | { resources?: { id: number; updated?: string; created?: string }[] }
        | undefined
      return (data?.resources ?? []).map((r) => ({
        path: `/galgame-resource/${r.id}`,
        lastmod: toIso(r.updated ?? r.created)
      }))
    },
    (json) => (unwrap(json) as { total?: number })?.total ?? 0
  )

// Galgame ratings — /api/galgame-rating/all.
const fetchRatings = (): Promise<KunDynamicRoute[]> =>
  fetchPaginated<KunDynamicRoute>(
    '/galgame-rating/all',
    (json) => {
      const data = unwrap(json) as
        | {
            ratingData?: { id: number; updated?: string; created?: string }[]
          }
        | undefined
      return (data?.ratingData ?? []).map((r) => ({
        path: `/galgame-rating/${r.id}`,
        lastmod: toIso(r.updated ?? r.created)
      }))
    },
    (json) => (unwrap(json) as { total?: number })?.total ?? 0
  )

// Series list — /api/galgame-series.
const fetchSeries = (): Promise<KunDynamicRoute[]> =>
  fetchPaginated<KunDynamicRoute>(
    '/galgame-series',
    (json) => {
      const data = unwrap(json) as
        | { series?: { id: number; updated?: string }[] }
        | undefined
      return (data?.series ?? []).map((s) => ({
        path: `/galgame-series/${s.id}`,
        lastmod: toIso(s.updated)
      }))
    },
    (json) => (unwrap(json) as { total?: number })?.total ?? 0
  )

// Tag list — /api/galgame-tag (no pagination total wrapper).
const fetchTags = async (): Promise<KunDynamicRoute[]> => {
  try {
    const res = await fetch(API + '/galgame-tag', { headers: baseHeaders })
    if (!res.ok) return []
    const data = unwrap(await res.json()) as
      | { tags?: { id: number; updated?: string }[] }
      | undefined
    return (data?.tags ?? []).map((t) => ({
      path: `/galgame-tag/${t.id}`,
      lastmod: toIso(t.updated)
    }))
  } catch {
    return []
  }
}

// Engine list — /api/galgame-engine; emit /galgame-engine/:name URLs.
const fetchEngines = async (): Promise<KunDynamicRoute[]> => {
  try {
    const res = await fetch(API + '/galgame-engine', { headers: baseHeaders })
    if (!res.ok) return []
    const data = unwrap(await res.json()) as
      | { id: number; name: string; updated?: string }[]
      | undefined
    return (Array.isArray(data) ? data : []).map((e) => ({
      path: `/galgame-engine/${encodeURIComponent(e.name)}`,
      lastmod: toIso(e.updated)
    }))
  } catch {
    return []
  }
}

// Official list — /api/galgame-official.
const fetchOfficials = (): Promise<KunDynamicRoute[]> =>
  fetchPaginated<KunDynamicRoute>(
    '/galgame-official',
    (json) => {
      const data = unwrap(json) as
        | { officials?: { id: number; name: string; updated?: string }[] }
        | undefined
      return (data?.officials ?? []).map((o) => ({
        path: `/galgame-official/${encodeURIComponent(o.name)}`,
        lastmod: toIso(o.updated)
      }))
    },
    (json) => (unwrap(json) as { total?: number })?.total ?? 0
  )

export const getKunDynamicRoutes = async (): Promise<KunDynamicRoute[]> => {
  // Run all top-level fetches in parallel — they're independent.
  // Per-endpoint failures are swallowed inside each helper, so one bad
  // service doesn't take down the whole sitemap build.
  const groups = await Promise.all([
    fetchTopics(),
    fetchGalgames(),
    fetchResources(),
    fetchRatings(),
    fetchSeries(),
    fetchTags(),
    fetchEngines(),
    fetchOfficials()
  ])
  return groups.flat()
}
