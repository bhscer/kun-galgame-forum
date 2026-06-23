<script setup lang="ts">
// Desktop sidebar icon rail: four groups (icon + label), each a RailGroup whose
// hover reveals a flyout of links to its right. Hover handling lives in RailGroup
// (kun-ui 2.1 useKunPointerMenu — safe-triangle + shared group). Mobile never
// renders this — it keeps the expanded drawer.
import { kunSidebarRail } from '~/constants/layout'
</script>

<template>
  <nav class="flex flex-col items-center gap-1">
    <KunLayoutSideBarRailGroup
      v-for="group in kunSidebarRail"
      :key="group.name"
      :group="group"
    />
  </nav>
</template>
