<script setup lang="ts">
import { kunLayoutExternalItem } from '~/constants/layout'

useKunSeoMeta({
  title: '子网站',
  description: kunLayoutExternalItem.map((site) => site.label).join(', ')
})
</script>

<template>
  <SitesContainer />
</template>
