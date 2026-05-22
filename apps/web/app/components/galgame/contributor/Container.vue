<script setup lang="ts">
// Read-only display of the galgame's contributors.
//
// Editing contributors is a wiki-owned concern — kungal shouldn't host
// add / remove UI for the same row that the wiki canonically governs.
// The previous version fetched `/galgame/:gid/contributors` and offered
// a DELETE button gated on creator/admin; both removed.
//
// Data source: the galgame detail's flat `contributor: UserBrief[]`
// already provided via the injected `galgame` (Galgame.vue's provide).
// Hover an avatar to see the contributor's name — click navigates to
// their user page (KunAvatar default).
const galgame = inject<GalgameDetail>('galgame')
const contributors = computed(() => galgame?.contributor ?? [])
</script>

<template>
  <div v-if="contributors.length" class="flex flex-wrap items-center gap-1">
    <KunTooltip
      v-for="user in contributors"
      :key="user.id"
      :text="user.name"
    >
      <KunAvatar :user="user" size="sm" />
    </KunTooltip>
  </div>
  <KunNull v-else description="暂无贡献者" />
</template>
