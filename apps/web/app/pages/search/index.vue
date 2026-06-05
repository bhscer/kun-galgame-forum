<script setup lang="ts">
// noindex on purpose: search-result pages are thin/duplicate content that
// shouldn't be indexed (Google's own guidance), and search now surfaces ALL
// games regardless of NSFW (the content_limit filter was removed) — so an
// indexable /search could leak R18 entries into search engines. R18 keeps out
// of the index at the page layer (here + the conditional noindex on R18 detail
// pages), NOT by filtering search results. moyu's /search already does this.
useKunDisableSeo('搜索 Galgame')
</script>

<template>
  <SearchContainer />
</template>
