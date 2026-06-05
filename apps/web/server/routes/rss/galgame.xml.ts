import { fetchKunApi, useKunFeed } from '../../utils/kunFeed'

interface GalgameRSSItem {
  id: number
  name: string
  banner: string
  user: { id: number; name: string; avatar: string }
  description: string
  created: string
}

// Cache the upstream API/DB round-trip (not the whole handler) so RSS readers
// polling every few minutes collapse into ONE upstream fetch per 5-min window.
// Caching the function (rather than defineCachedEventHandler) means only a
// SUCCESSFUL fetch is cached — a transient API failure can't pin an empty feed
// for the whole TTL.
const fetchGalgameRssItems = defineCachedFunction(
  () => fetchKunApi<GalgameRSSItem[]>('/rss/galgame'),
  { maxAge: 60 * 5, name: 'rss-galgame', getKey: () => 'all' }
)

export default defineEventHandler(async (event) => {
  const baseUrl = useRuntimeConfig().public.KUN_GALGAME_URL || ''
  const feed = useKunFeed(baseUrl, 'galgame')

  // Resilient: a slow/unreachable API must not turn an RSS poll into an
  // unhandled 500 (those flooded prod logs). Serve an empty-but-valid feed and
  // retry on the next poll.
  try {
    const items = await fetchGalgameRssItems()
    for (const g of items) {
      feed.addItem({
        link: `${baseUrl}/galgame/${g.id}`,
        title: g.name,
        date: new Date(g.created),
        description: g.description,
        image: g.banner,
        author: [
          {
            name: g.user.name,
            link: `${baseUrl}/user/${g.user.id}/info`
          }
        ]
      })
    }
  } catch (error) {
    console.error('[rss] galgame feed upstream failed:', error)
  }

  setHeader(event, 'Content-Type', 'application/xml')
  return feed.rss2()
})
