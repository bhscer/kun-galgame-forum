<script setup lang="ts">
import { kunSubSites } from '~/constants/layout'

useKunSeoMeta({
  title: '子网站',
  description: kunSubSites.map((site) => site.short).join(', ')
})
</script>

<template>
  <SitesContainer />
</template>
