<script setup lang="ts">
const siteList = [
  { name: '补丁站', href: kungal.domain.patch },
  { name: '表情包', href: kungal.domain.sticker },
  { name: '导航网站', href: kungal.domain.nav },
  { name: '开发文档', href: kungal.domain.doc }
]
</script>

<template>
  <!-- Simple Card-less footer tuned for the narrow right rail: everything wraps
       + centers. No top divider — it sits at the bottom of the rail, spaced off
       the info above by the rail's flex gap. -->
  <footer class="text-default-500 space-y-2.5 text-center text-xs">

    <div class="flex flex-wrap justify-center gap-x-3 gap-y-1">
      <KunLink
        v-for="(site, index) in siteList"
        :key="index"
        size="sm"
        :to="site.href"
        target="_blank"
      >
        {{ site.name }}
      </KunLink>
    </div>

    <div class="flex flex-wrap items-center justify-center gap-1.5">
      <a
        :href="kungal.domain.main"
        target="_blank"
        class="flex items-center"
        :aria-label="kungal.titleShort"
      >
        <KunImage class="size-5 rounded-md" src="/favicon.webp" :alt="kungal.titleShort" />
      </a>
      <span>© 2023 ~ Forever! {{ kungal.titleShort }}</span>
    </div>

    <div class="flex flex-wrap items-center justify-center gap-1">
      <KunLink
        :to="kungal.github"
        size="sm"
        target="_blank"
        rel="noopener noreferrer"
      >
        GitHub
      </KunLink>
      <span>开源 · v{{ useRuntimeConfig().public.KUN_VISUAL_NOVEL_VERSION }}</span>
    </div>
  </footer>
</template>
