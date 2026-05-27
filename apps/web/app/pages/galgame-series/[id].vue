<script setup lang="ts">
// Series detail intentionally does NOT show revision history — series
// membership changes are recorded as galgame-side revisions (each
// affected galgame gets its own `series_id` change), so a per-series
// revision feed would be empty/misleading. See K-PR series-revision
// design note for context.
const route = useRoute()

const seriesId = computed(() => {
  return Number((route.params as { id: string }).id)
})

const { data } = await useKunFetch<GalgameSeriesDetail>(`/galgame-series/${seriesId.value}`, {
  method: 'GET',
  query: { seriesId: seriesId.value }
})

if (data.value) {
  if (data.value.isNSFW) {
    useKunDisableSeo(data.value.name)
  } else {
    useKunSeoMeta({
      title: `${data.value.name} 系列下载资源`,
      description: data.value.description
    })
  }
} else {
  useKunDisableSeo('未找到 Galgame 系列')
}
</script>

<template>
  <div class="contents">
    <GalgameSeriesDetail :data="data" v-if="data" />
    <KunNull v-else description="未找到这个 Galgame 系列" />
  </div>
</template>
