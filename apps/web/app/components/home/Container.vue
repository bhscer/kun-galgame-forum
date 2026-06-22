<script setup lang="ts">
import { useIntersectionObserver } from '@vueuse/core'

// Home page = the activity feed, split into five tabs. Each tab is a server-side
// bucket of activity types (GET /activity/tab?tab=…); "全部" is every type EXCEPT
// galgame resources, which get their own 资源 tab so the main stream isn't
// drowned by download spam. Keep tab order/values in lock-step with the backend
// homeTabTypes map (activity_service.go). Cards reuse the /activity card style;
// no KunCard wrapper here (the feed is the page).
const HOME_FEED_TABS = [
  { value: 'all', textValue: '全部', icon: 'lucide:layers' },
  { value: 'topic', textValue: '话题', icon: 'icon-park-outline:topic' },
  { value: 'galgame', textValue: 'Galgame', icon: 'lucide:gamepad-2' },
  { value: 'resource', textValue: '资源', icon: 'lucide:box' },
  { value: 'others', textValue: '其他', icon: 'lucide:layout-grid' }
]

const settings = usePersistSettingsStore()
const activeTab = ref('all')

const items = ref<ActivityItem[]>([])
const cursor = ref('')
const hasMore = ref(true)
const isLoadingMore = ref(false)

// First page is SSR-rendered (SEO, no spinner). The reactive query re-fetches
// when the tab or the showNoResource toggle changes; the watch re-seeds the
// accumulator from whichever first page is current.
const { data, status } = await useKunFetch<{
  items: ActivityItem[]
  nextCursor: string
}>('/activity/tab', {
  method: 'GET',
  query: computed(() => ({
    tab: activeTab.value,
    limit: 30,
    showNoResource: settings.showKUNGalgameNoResource
  }))
})

watch(
  data,
  (page) => {
    if (!page) return
    items.value = page.items
    cursor.value = page.nextCursor
    hasMore.value = !!page.nextCursor
  },
  { immediate: true }
)

// On a tab switch, drop the stale cursor immediately so an in-flight infinite
// scroll can't append the OLD tab's next page onto the NEW tab (the data watch
// re-seeds once the new first page arrives; old items stay until then — no flash).
watch(activeTab, () => {
  cursor.value = ''
  hasMore.value = true
})

const loadMore = async () => {
  if (isLoadingMore.value || !hasMore.value || !cursor.value) return
  isLoadingMore.value = true
  const next = await kunFetch<{ items: ActivityItem[]; nextCursor: string }>(
    '/activity/tab',
    {
      method: 'GET',
      query: {
        tab: activeTab.value,
        limit: 30,
        cursor: cursor.value,
        showNoResource: settings.showKUNGalgameNoResource
      }
    }
  )
  isLoadingMore.value = false
  if (!next) return
  items.value.push(...next.items)
  cursor.value = next.nextCursor
  hasMore.value = !!next.nextCursor
}

// Auto-load near the bottom (VueUse no-ops on SSR); the button is the fallback.
const sentinel = ref<HTMLElement | null>(null)
useIntersectionObserver(
  sentinel,
  ([entry]) => {
    if (entry?.isIntersecting) loadMore()
  },
  { rootMargin: '400px' }
)
</script>

<template>
  <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
    <!-- Mobile: horizontal underline tabs on top. -->
    <div class="sm:hidden">
      <KunTab
        v-model="activeTab"
        :items="HOME_FEED_TABS"
        variant="underlined"
        color="primary"
        full-width
      />
    </div>

    <!-- Desktop: vertical underline tab rail on the left (like the settings
         panel). Sticky so it stays put while the center feed scrolls. -->
    <div class="sticky top-20 hidden shrink-0 self-start sm:block sm:w-28">
      <KunTab
        v-model="activeTab"
        :items="HOME_FEED_TABS"
        orientation="vertical"
        variant="underlined"
        color="primary"
        full-width
      />
    </div>

    <div class="min-w-0 flex-1">
      <KunNull
        v-if="status !== 'pending' && !items.length"
        description="暂无动态"
      />

      <div v-else class="divide-default-200/60 divide-y">
        <div
          v-for="activity in items"
          :key="activity.uniqueId"
          class="py-5 first:pt-0 last:pb-0"
        >
          <ActivityCard :activity="activity" />
        </div>
      </div>

      <div v-if="items.length" ref="sentinel" class="flex justify-center pt-4">
        <KunButton
          v-if="hasMore"
          variant="light"
          :loading="isLoadingMore"
          @click="loadMore"
        >
          加载更多
        </KunButton>
        <span v-else class="text-default-400 text-sm">没有更多动态了</span>
      </div>
    </div>

    <!-- Right rail (desktop only, ≥lg): carousel · 使用提示 at the top, footer
         pinned to the bottom. Sticky + viewport-tall so it stays fixed while the
         center feed scrolls; fixed width keeps the feed the focus (~65-70%). -->
    <aside
      class="sticky top-20 hidden h-[calc(100dvh-6rem)] shrink-0 flex-col self-start lg:flex lg:w-72 xl:w-80"
    >
      <div class="space-y-4">
        <HomeCarousel />
        <HomeAsideHelp />
      </div>
      <HomeFooter class="mt-auto" />
    </aside>
  </div>
</template>
