<script setup lang="ts">
// U2 screenshot gallery — read-only display on the detail page.
//
// Clicking a thumbnail opens a connected KunLightboxGallery (prev/next + zoom
// across the whole gallery in one overlay), the same lightbox the banner uses.
//
// Per-rating SFW gate — two independent axes, both filtered client-side so the
// editor's replace-all submit never loses rows (see GalleryFilter.vue):
//   色情 (sexual)   — global NSFW mode shows every level; SFW shows only the
//                     levels the viewer opted into.
//   暴力 (violence) — independent per-level opt-in (default off, behind a
//                     warning); the NSFW mode does NOT unlock it.
// An image shows iff BOTH its sexual and violence levels are permitted. The
// filter control stays reachable even when everything is hidden, so a fully
// rated gallery can still be revealed. Selections persist and the NSFW toggle
// reloads, so SSR and client agree.
//
// Image source: galgameImageSrc resolves each row's cdn_url (kungal injects it
// via rewriteBanners) and falls back to the /image/<hash> token (web 302
// middleware) when a row lacks cdn_url — so a screenshot still shows from
// image_hash alone instead of being silently dropped.
const props = defineProps<{
  screenshots: GalgameScreenshot[]
}>()

const {
  showKUNGalgameContentLimit,
  showKUNGalgameGallerySexualLevels: sexualLevels,
  showKUNGalgameGalleryViolenceLevels: violenceLevels
} = storeToRefs(usePersistSettingsStore())

const showNsfw = computed(
  () =>
    showKUNGalgameContentLimit.value === 'nsfw' ||
    showKUNGalgameContentLimit.value === 'all'
)

// 色情: NSFW reveals every level; otherwise unrated (0) + opted-in levels.
// 暴力: unrated (0) + opted-in levels only, independent of the NSFW mode.
const sexualOk = (s: GalgameScreenshot) =>
  showNsfw.value || s.sexual === 0 || sexualLevels.value.includes(s.sexual)
const violenceOk = (s: GalgameScreenshot) =>
  s.violence === 0 || violenceLevels.value.includes(s.violence)

const allShots = computed(() =>
  [...(props.screenshots ?? [])].filter((s) => !!s.image_hash)
)

const sorted = computed(() =>
  allShots.value
    .filter((s) => sexualOk(s) && violenceOk(s))
    .sort((a, b) => {
      if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
      return a.image_hash.localeCompare(b.image_hash)
    })
)

const hiddenCount = computed(() => allShots.value.length - sorted.value.length)

// Only surface the filter when there's something to gate — a gallery of purely
// unrated screenshots needs no control.
const hasRated = computed(() =>
  allShots.value.some((s) => s.sexual >= 1 || s.violence >= 1)
)
</script>

<template>
  <div v-if="allShots.length" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-2">
      <KunHeader name="画廊" description="该 Galgame 的截图 / CG 集" scale="h3" />
      <GalgameGalleryFilter
        v-if="hasRated"
        :show-nsfw="showNsfw"
        :hidden-count="hiddenCount"
      />
    </div>

    <KunLightboxGallery v-if="sorted.length">
      <div
        class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
      >
        <KunLightboxGalleryItem
          v-for="s in sorted"
          :key="s.image_hash"
          :src="galgameImageSrc(s)"
          :alt="s.caption || ''"
          :wrap="false"
          v-slot="{ open }"
        >
          <button
            type="button"
            class="group hover:ring-primary focus:ring-primary relative block w-full overflow-hidden rounded-lg ring-1 ring-transparent transition-all focus:outline-none"
            :aria-label="s.caption || '查看截图'"
            @click="open"
          >
            <KunImage
              :src="galgameImageSrc(s)"
              :alt="s.caption || ''"
              loading="lazy"
              class-name="aspect-video w-full cursor-zoom-in object-cover transition-transform duration-200 group-hover:scale-105"
            />
            <div
              v-if="s.caption"
              class="absolute right-0 bottom-0 left-0 truncate bg-black/50 px-2 py-1 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100"
            >
              {{ s.caption }}
            </div>
          </button>
        </KunLightboxGalleryItem>
      </div>
    </KunLightboxGallery>

    <KunNull
      v-else
      :description="`${hiddenCount} 张图片已按分级隐藏，点击「分级筛选」调整`"
    />
  </div>
</template>
