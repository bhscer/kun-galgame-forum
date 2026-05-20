<script setup lang="ts">
// Cover candidate-set editor (U2 K-PR3b). Replaces the old single-banner
// upload — covers[sort_order=0] is the "pinned banner" (= effective
// banner head image); other rows are alternate candidates. Reorder by
// changing sort_order; pin = make sort_order=0 (the wiki has a partial
// unique index so two rows with sort_order=0 will be rejected, hence
// the explicit "钉为封面" action atomically demotes the old pinned).
//
// Upload uses K-PR3a's /image/galgame proxy → returns hash + URL; we
// push a new row with sort_order set to (max+1) and unrated sexual/
// violence (admin or future per-image gate may rate later).
//
// Footer.vue strips `cdn_url` (server-injected preview) from the wire
// payload — wiki doesn't accept it on write.
import type { KunSelectOption } from '~/components/kun/select/type'

const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const covers = computed<GalgameCover[]>({
  get: () => galgamePR.value[0]?.covers ?? [],
  set: (v) => {
    if (galgamePR.value[0]) galgamePR.value[0].covers = v
  }
})

const sorted = computed(() =>
  [...covers.value].sort((a, b) => {
    if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
    return a.image_hash.localeCompare(b.image_hash)
  })
)

const ratingOptions: KunSelectOption[] = [
  { value: 0, label: '未评定' },
  { value: 1, label: '轻' },
  { value: 2, label: '中' },
  { value: 3, label: '高' }
]

const isUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const handlePickFile = () => {
  fileInput.value?.click()
}

const handleFile = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  // Reset so the same file can be picked twice if a transient error
  // occurred (browser otherwise dedupes by identity and the event
  // doesn't refire).
  input.value = ''
  if (!file) return

  isUploading.value = true
  const res = await uploadGalgameImage(file, 'galgame_banner', file.name)
  isUploading.value = false
  if (!res) return // kunFetch already toasted the wiki error message

  // De-dupe: same hash on the same galgame is meaningless; image_service
  // deduplicates by content so re-uploading the same file returns the
  // same hash. Surface a friendly note instead of pushing a phantom row.
  if (covers.value.some((c) => c.image_hash === res.hash)) {
    useMessage('该封面已在列表中', 'warn')
    return
  }

  const maxOrder = covers.value.reduce(
    (m, c) => (c.sort_order > m ? c.sort_order : m),
    -1
  )
  covers.value = [
    ...covers.value,
    {
      image_hash: res.hash,
      sort_order: maxOrder + 1,
      sexual: 0,
      violence: 0,
      source: '',
      source_key: '',
      cdn_url: res.url
    }
  ]
}

// Atomic pin: set picked → 0, demote old 0-row to (max+1). Mirrors the
// wiki's "Pin new banner" transaction so the partial unique index is
// never violated on submit. Idempotent: pinning the already-pinned row
// is a no-op.
const handlePin = (target: GalgameCover) => {
  if (target.sort_order === 0) return
  const maxOrder = covers.value.reduce(
    (m, c) => (c.sort_order > m ? c.sort_order : m),
    0
  )
  covers.value = covers.value.map((c) => {
    if (c.image_hash === target.image_hash) return { ...c, sort_order: 0 }
    if (c.sort_order === 0) return { ...c, sort_order: maxOrder + 1 }
    return c
  })
}

const handleRemove = (hash: string) => {
  covers.value = covers.value.filter((c) => c.image_hash !== hash)
}
</script>

<template>
  <div class="space-y-3">
    <h2 class="text-xl">封面</h2>
    <p class="text-default-500 text-sm">
      添加候选封面图;其中 <span class="text-primary">已钉为封面</span>
      的一张作为详情页头图(`effective_banner`),其余为备选。
      提交后将整组替换该 Galgame 的封面集。
    </p>

    <div v-if="sorted.length" class="space-y-2">
      <div
        v-for="cover in sorted"
        :key="cover.image_hash"
        class="border-default-200 flex flex-col gap-3 rounded-lg border p-3 sm:flex-row"
      >
        <KunImage
          v-if="cover.cdn_url"
          :src="cover.cdn_url"
          loading="lazy"
          class-name="h-24 w-40 shrink-0 rounded object-cover"
        />
        <div class="flex-1 space-y-2">
          <div class="flex flex-wrap items-center gap-2">
            <KunBadge
              v-if="cover.sort_order === 0"
              color="primary"
              variant="flat"
            >
              已钉为封面
            </KunBadge>
            <KunButton
              v-else
              size="sm"
              variant="flat"
              @click="handlePin(cover)"
            >
              钉为封面
            </KunButton>
            <span class="text-default-400 text-xs">
              sort_order = {{ cover.sort_order }}
            </span>
          </div>
          <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
            <KunSelect
              v-model="cover.sexual"
              label="性内容评级"
              :options="ratingOptions"
            />
            <KunSelect
              v-model="cover.violence"
              label="暴力评级"
              :options="ratingOptions"
            />
          </div>
        </div>
        <KunButton
          :is-icon-only="true"
          variant="flat"
          color="danger"
          @click="handleRemove(cover.image_hash)"
        >
          <KunIcon name="lucide:trash-2" />
        </KunButton>
      </div>
    </div>

    <input
      ref="fileInput"
      type="file"
      accept="image/jpeg,image/png,image/webp"
      class="hidden"
      @change="handleFile"
    />
    <KunButton variant="flat" :loading="isUploading" @click="handlePickFile">
      <KunIcon name="lucide:plus" />
      添加封面
    </KunButton>
  </div>
</template>
