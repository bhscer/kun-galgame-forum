<script setup lang="ts">
// Screenshot gallery editor (U2 K-PR3b). Each row references an image
// in image_service by hash; rows ordered by sort_order. Distinct from
// covers — no pinned/banner concept, just an ordered gallery with
// optional captions.
//
// Upload preset is galgame_screenshot (image_service may apply a
// different size budget than the banner preset). Footer strips
// `cdn_url` from the wire payload (read-only preview field).
import type { KunSelectOption } from '~/components/kun/select/type'

const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const screenshots = computed<GalgameScreenshot[]>({
  get: () => galgamePR.value[0]?.screenshots ?? [],
  set: (v) => {
    if (galgamePR.value[0]) galgamePR.value[0].screenshots = v
  }
})

const sorted = computed(() =>
  [...screenshots.value].sort((a, b) => {
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
  input.value = ''
  if (!file) return

  isUploading.value = true
  const res = await uploadGalgameImage(file, 'galgame_screenshot', file.name)
  isUploading.value = false
  if (!res) return

  if (screenshots.value.some((s) => s.image_hash === res.hash)) {
    useMessage('该截图已在列表中', 'warn')
    return
  }

  const maxOrder = screenshots.value.reduce(
    (m, s) => (s.sort_order > m ? s.sort_order : m),
    -1
  )
  screenshots.value = [
    ...screenshots.value,
    {
      image_hash: res.hash,
      sort_order: maxOrder + 1,
      caption: '',
      sexual: 0,
      violence: 0,
      source: '',
      source_key: '',
      cdn_url: res.url
    }
  ]
}

const handleRemove = (hash: string) => {
  screenshots.value = screenshots.value.filter((s) => s.image_hash !== hash)
}

// Shift up / down by swapping sort_order with the neighbour. v1
// keeps reorder simple (no drag-and-drop per §10 #3).
const handleShift = (target: GalgameScreenshot, dir: -1 | 1) => {
  const idx = sorted.value.findIndex(
    (s) => s.image_hash === target.image_hash
  )
  const neighbour = sorted.value[idx + dir]
  if (!neighbour) return
  const a = target.sort_order
  const b = neighbour.sort_order
  screenshots.value = screenshots.value.map((s) => {
    if (s.image_hash === target.image_hash) return { ...s, sort_order: b }
    if (s.image_hash === neighbour.image_hash) return { ...s, sort_order: a }
    return s
  })
}
</script>

<template>
  <div class="space-y-3">
    <h2 class="text-xl">画廊 / 截图</h2>
    <p class="text-default-500 text-sm">
      游戏截图 / CG / 立绘等。按 sort_order 排序。提交后将整组替换。
    </p>

    <div v-if="sorted.length" class="space-y-2">
      <div
        v-for="(shot, idx) in sorted"
        :key="shot.image_hash"
        class="border-default-200 flex flex-col gap-3 rounded-lg border p-3 sm:flex-row"
      >
        <KunImage
          v-if="shot.cdn_url"
          :src="shot.cdn_url"
          loading="lazy"
          class-name="h-24 w-40 shrink-0 rounded object-cover"
        />
        <div class="flex-1 space-y-2">
          <div class="flex flex-wrap items-center gap-2">
            <KunButton
              size="sm"
              variant="flat"
              :is-icon-only="true"
              :disabled="idx === 0"
              @click="handleShift(shot, -1)"
            >
              <KunIcon name="lucide:chevron-up" />
            </KunButton>
            <KunButton
              size="sm"
              variant="flat"
              :is-icon-only="true"
              :disabled="idx === sorted.length - 1"
              @click="handleShift(shot, 1)"
            >
              <KunIcon name="lucide:chevron-down" />
            </KunButton>
            <span class="text-default-400 text-xs">
              sort_order = {{ shot.sort_order }}
            </span>
          </div>
          <KunInput v-model="shot.caption" placeholder="说明文字 (可选)" />
          <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
            <KunSelect
              v-model="shot.sexual"
              label="性内容评级"
              :options="ratingOptions"
            />
            <KunSelect
              v-model="shot.violence"
              label="暴力评级"
              :options="ratingOptions"
            />
          </div>
        </div>
        <KunButton
          :is-icon-only="true"
          variant="flat"
          color="danger"
          @click="handleRemove(shot.image_hash)"
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
      添加截图
    </KunButton>
  </div>
</template>
