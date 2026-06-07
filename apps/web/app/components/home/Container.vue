<script setup lang="ts">
// Global "显示没有下载资源的 Galgame" preference (cookie-persisted, SSR-safe).
// Off (default) drops resource-less galgames' creation rows from the feed too.
const settings = usePersistSettingsStore()
const [{ data }, { data: activityData }] = await Promise.all([
  useKunFetch<{
    galgames: HomeGalgame[]
    topics: HomeTopic[]
  }>('/home'),
  useKunFetch<{
    items: ActivityItem[]
    total: number
  }>('/activity', {
    query: {
      page: 1,
      limit: 20,
      type: 'all',
      showNoResource: settings.showKUNGalgameNoResource
    }
  })
])

const activities = computed(() => activityData.value?.items || [])
</script>

<template>
  <div class="flex flex-col justify-between gap-3 rounded-lg lg:flex-row">
    <div class="w-full space-y-3">
      <HomeCarousel />

      <HomeTopicContainer v-if="data" :topics="data.topics" />

      <HomeGalgameContainer v-if="data" :galgames="data.galgames" />

      <HomeFooter />
    </div>

    <div class="w-full shrink-0 space-y-3 lg:w-72">
      <HomeAsideRecent v-if="activities.length" :activities="activities" />

      <HomeAsideHelp />
    </div>
  </div>
</template>
