<script setup lang="ts">
// U2 screenshot gallery — read-only display on the detail page.
//
// v1 spec (deliberately minimal): grid thumbnails sorted by sort_order
// then by ImageHash for stable ordering, click → KunModal full-screen
// preview of the picked screenshot. No lightbox keyboard nav, no zoom,
// no carousel. Empty list → render nothing (the section disappears so
// older galgames with no screenshots don't get an empty placeholder).
//
// Image source: each row's `cdn_url` (kungal injects it via
// rewriteBanners). If a row somehow lacks `cdn_url` we skip it rather
// than show a broken image.
const props = defineProps<{
  screenshots: GalgameScreenshot[]
}>()

const sorted = computed(() =>
  [...(props.screenshots ?? [])]
    .filter((s) => !!s.cdn_url)
    .sort((a, b) => {
      if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
      return a.image_hash.localeCompare(b.image_hash)
    })
)

const isOpen = ref(false)
const activeHash = ref<string | null>(null)
const active = computed(() =>
  sorted.value.find((s) => s.image_hash === activeHash.value) ?? null
)

const open = (hash: string) => {
  activeHash.value = hash
  isOpen.value = true
}
</script>

<template>
  <div v-if="sorted.length" class="space-y-3">
    <KunHeader
      name="画廊"
      description="该 Galgame 的截图 / CG 集"
      scale="h3"
    />

    <div
      class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
    >
      <button
        v-for="s in sorted"
        :key="s.image_hash"
        type="button"
        class="group hover:ring-primary focus:ring-primary relative overflow-hidden rounded-lg ring-1 ring-transparent transition-all focus:outline-none"
        :aria-label="s.caption || '查看截图'"
        @click="open(s.image_hash)"
      >
        <KunImage
          :src="s.cdn_url!"
          :alt="s.caption || ''"
          loading="lazy"
          class-name="aspect-video w-full object-cover transition-transform duration-200 group-hover:scale-105"
        />
        <div
          v-if="s.caption"
          class="absolute bottom-0 left-0 right-0 truncate bg-black/50 px-2 py-1 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100"
        >
          {{ s.caption }}
        </div>
      </button>
    </div>

    <KunModal v-model:modal-value="isOpen" inner-class-name="max-w-4xl">
      <div v-if="active" class="space-y-2">
        <KunImage
          :src="active.cdn_url!"
          :alt="active.caption || ''"
          class-name="max-h-[80vh] w-full object-contain"
        />
        <p v-if="active.caption" class="text-default-600 text-center text-sm">
          {{ active.caption }}
        </p>
      </div>
    </KunModal>
  </div>
</template>
