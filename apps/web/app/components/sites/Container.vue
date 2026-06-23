<script setup lang="ts">
// Standalone "子网站" page: the KUN Galgame family of sites (patch station,
// sticker pack, dev docs, nav site). Sourced from the same kunLayoutExternalItem
// the sidebar used to embed; promoted to its own page in the layout refactor.
import { kunLayoutExternalItem } from '~/constants/layout'

// Tag outbound links with utm_source=<current domain>, like friend links.
const utmLink = useUtmLink()
</script>

<template>
  <div class="space-y-6 pb-12">
    <KunHeader
      name="子网站"
      description="鲲 Galgame 旗下的其他站点与资源, 均为我们自己维护, 欢迎逛逛~"
    />

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <KunCard
        v-for="site in kunLayoutExternalItem"
        :key="site.name"
        :is-transparent="false"
        :is-hoverable="true"
        :href="utmLink(site.router!)"
        target="_blank"
        rel="noopener noreferrer"
        content-class="space-y-2"
      >
        <div class="flex items-center gap-2">
          <KunIcon
            v-if="site.icon"
            :name="site.icon"
            class="text-primary text-2xl"
          />
          <span class="text-lg font-bold">{{ site.label }}</span>
          <KunChip v-if="site.hint" color="primary" size="sm">
            {{ site.hint }}
          </KunChip>
        </div>
        <div class="text-default-500 text-sm break-all">
          {{ site.router }}
        </div>
      </KunCard>
    </div>
  </div>
</template>
