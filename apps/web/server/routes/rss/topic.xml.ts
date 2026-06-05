import { fetchKunApi, useKunFeed } from '../../utils/kunFeed'

interface TopicRSSItem {
  id: number
  title: string
  description: string
  userId: number
  userName: string
  created: string
}

// Cache the upstream API/DB round-trip (not the whole handler) so RSS readers
// polling every few minutes collapse into ONE upstream fetch per 5-min window.
// Caching the function (rather than defineCachedEventHandler) means only a
// SUCCESSFUL fetch is cached — a transient API failure can't pin an empty feed
// for the whole TTL.
const fetchTopicRssItems = defineCachedFunction(
  () => fetchKunApi<TopicRSSItem[]>('/rss/topic'),
  { maxAge: 60 * 5, name: 'rss-topic', getKey: () => 'all' }
)

export default defineEventHandler(async (event) => {
  const baseUrl = useRuntimeConfig().public.KUN_GALGAME_URL || ''
  const feed = useKunFeed(baseUrl, 'topic')

  // Resilient: a slow/unreachable API must not turn an RSS poll into an
  // unhandled 500 (those flooded prod logs). Serve an empty-but-valid feed and
  // retry on the next poll.
  try {
    const topics = await fetchTopicRssItems()
    for (const t of topics) {
      feed.addItem({
        link: `${baseUrl}/topic/${t.id}`,
        title: t.title,
        date: new Date(t.created),
        description: t.description,
        author: [
          {
            name: t.userName,
            link: `${baseUrl}/user/${t.userId}/info`
          }
        ]
      })
    }
  } catch (error) {
    console.error('[rss] topic feed upstream failed:', error)
  }

  setHeader(event, 'Content-Type', 'application/xml')
  return feed.rss2()
})
