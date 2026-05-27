<script setup lang="ts">
const route = useRoute()
const { role } = usePersistUserStore()

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

    <GalgameRevisionList
      v-if="data"
      entity="series"
      :id="seriesId"
      :entity-label="`系列「${data.name}」`"
      :can-revert="role >= 2"
    />
  </div>
</template>
