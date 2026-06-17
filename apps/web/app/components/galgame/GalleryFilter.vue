<script setup lang="ts">
// Per-rating gallery filter (galgame detail 画廊). Two independent axes:
//   色情 (sexual)   — gated by the global NSFW mode: NSFW shows every level,
//                     SFW reveals only the levels opted-in here.
//   暴力 (violence) — ALWAYS an explicit per-level opt-in (default off), shown
//                     behind a prominent confirm. The NSFW mode does NOT unlock
//                     it (sex and gore are separate sensitivities).
// An image needs BOTH its sexual and violence levels permitted to show — the
// actual hide/show runs in Gallery.vue; this component only edits the persisted
// level sets (remembered across pages/sessions via the settings store).
import type { Ref } from 'vue'

defineProps<{ showNsfw: boolean; hiddenCount: number }>()

const {
  showKUNGalgameGallerySexualLevels: sexualLevels,
  showKUNGalgameGalleryViolenceLevels: violenceLevels
} = storeToRefs(usePersistSettingsStore())

const LEVELS = [
  { value: 1, label: '轻' },
  { value: 2, label: '中' },
  { value: 3, label: '高' }
]

const toggle = (arr: Ref<number[]>, level: number) => {
  arr.value = arr.value.includes(level)
    ? arr.value.filter((l) => l !== level)
    : [...arr.value, level]
}

// Dedicated handlers so the template never has to pass a ref (templates
// auto-unwrap refs, which would hand `toggle` a plain array).
const toggleSexual = (level: number) => toggle(sexualLevels, level)

// Violence: turning it ON from the all-off state pops a confirm first; once any
// violence level is on (or persisted from a prior session), further toggles
// don't re-prompt.
const warnOpen = ref(false)
const pendingLevel = ref<number | null>(null)

const onViolence = (level: number) => {
  const enabling = !violenceLevels.value.includes(level)
  if (enabling && violenceLevels.value.length === 0) {
    pendingLevel.value = level
    warnOpen.value = true
    return
  }
  toggle(violenceLevels, level)
}

const confirmViolence = () => {
  if (pendingLevel.value !== null) toggle(violenceLevels, pendingLevel.value)
  pendingLevel.value = null
  warnOpen.value = false
}
</script>

<template>
  <KunPopover position="bottom-end" inner-class="w-72 p-3">
    <template #trigger>
      <KunButton variant="flat" size="sm">
        <KunIcon name="lucide:funnel" />
        分级筛选
        <KunChip v-if="hiddenCount" size="sm" color="warning" variant="flat">
          隐藏 {{ hiddenCount }}
        </KunChip>
      </KunButton>
    </template>

    <div class="space-y-4">
      <!-- 色情 -->
      <div class="space-y-2">
        <p class="text-default-700 text-sm font-medium">色情评级</p>
        <p v-if="showNsfw" class="text-default-500 text-xs">
          NSFW 模式已显示全部色情内容。
        </p>
        <div v-else class="flex flex-wrap gap-3">
          <KunCheckBox
            v-for="lv in LEVELS"
            :id="`gal-sexual-${lv.value}`"
            :key="lv.value"
            type="single"
            :model-value="sexualLevels.includes(lv.value)"
            :label="lv.label"
            @update:model-value="() => toggleSexual(lv.value)"
          />
        </div>
      </div>

      <KunDivider />

      <!-- 暴力 -->
      <div class="space-y-2">
        <div class="flex items-center gap-1.5">
          <KunIcon name="lucide:triangle-alert" class="text-danger-500" />
          <p class="text-default-700 text-sm font-medium">暴力评级</p>
        </div>
        <p class="text-danger-600 text-xs">
          暴力 / 血腥内容可能引起不适,默认隐藏,确认后才显示。
        </p>
        <div class="flex flex-wrap gap-3">
          <KunCheckBox
            v-for="lv in LEVELS"
            :id="`gal-violence-${lv.value}`"
            :key="lv.value"
            type="single"
            :model-value="violenceLevels.includes(lv.value)"
            :label="lv.label"
            @update:model-value="() => onViolence(lv.value)"
          />
        </div>
      </div>
    </div>
  </KunPopover>

  <KunModal v-model="warnOpen" role="alertdialog" inner-class-name="max-w-sm">
    <div class="space-y-4 text-center">
      <KunIcon
        name="lucide:triangle-alert"
        class="text-danger-500 mx-auto text-4xl"
      />
      <div class="space-y-1">
        <p class="text-lg font-semibold">显示暴力内容?</p>
        <p class="text-default-600 text-sm">
          画廊中存在带有暴力 / 血腥评级的图片,可能引起不适。确认要显示这类内容吗?
        </p>
      </div>
      <div class="flex justify-center gap-2">
        <KunButton variant="flat" color="default" @click="warnOpen = false">
          取消
        </KunButton>
        <KunButton color="danger" @click="confirmViolence">确认显示</KunButton>
      </div>
    </div>
  </KunModal>
</template>
