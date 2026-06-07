<script setup lang="ts">
import { storeToRefs } from 'pinia'

// Filters are URL-backed (useGalgameFilters / useRouteQuery). The refs
// double as the fetch query — URL key === BE query key, so no remapping.
// useKunFetch watches the query refs, so a chip toggle (which writes the
// URL) re-fetches; browser back/forward re-fetches too.
const {
  page,
  limit,
  type,
  language,
  platform,
  sortField,
  sortOrder,
  releasedFrom,
  releasedTo,
  releasedMonths,
  includeProviders,
  excludeOnlyProviders,
  minRatingCount,
  minRating
} = useGalgameFilters()

// Global "显示没有下载资源的 Galgame" preference (cookie-persisted settings
// store, SSR-safe). Off (default) hides resource-less galgames. Passed in the
// query so toggling it re-fetches.
const { showKUNGalgameNoResource } = storeToRefs(usePersistSettingsStore())

const { data, status } = await useKunFetch<{
  galgames: GalgameCard[]
  total: number
}>(`/galgame`, {
  method: 'GET',
  query: {
    page,
    limit,
    type,
    language,
    platform,
    sortField,
    sortOrder,
    releasedFrom,
    releasedTo,
    releasedMonths,
    includeProviders,
    excludeOnlyProviders,
    minRatingCount,
    minRating,
    showNoResource: showKUNGalgameNoResource
  }
})
</script>

<template>
  <div class="flex flex-col gap-3">
    <template v-if="data">
      <KunCard class-name="z-10" :is-hoverable="false" :is-transparent="false">
        <KunHeader name="Galgame 资源 Wiki">
          <template #endContent>
            <GalgameCardNav :is-show-advanced="true" />
          </template>

          <template #description>
            <p class="text-default-500">
              Galgame 资源页面, 提供各类 Galgame 下载。我们不是资源的提供者,
              我们只是资源的指路人。
            </p>
          </template>
        </KunHeader>
      </KunCard>

      <KunLoading :loading="status === 'pending'">
        <GalgameCard v-if="data.galgames" :galgames="data.galgames" />
      </KunLoading>

      <KunCard
        :is-hoverable="false"
        :is-transparent="false"
        content-class="gap-3"
      >
        <KunPagination
          v-model:current-page="page"
          :total-page="Math.ceil(data.total / limit)"
          :is-loading="status === 'pending'"
        />
      </KunCard>
    </template>
  </div>
</template>
