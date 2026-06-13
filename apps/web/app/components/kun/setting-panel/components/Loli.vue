<script setup lang="ts">
import { getLoli } from './getLoli'

// ── Containing the 差分图 ──────────────────────────────────────────────
// The layer JSON (public/ren.json) places every part on the ORIGINAL game
// canvas. Across all selectable parts the assembled character occupies the
// rectangle (137,323)→(504,925) — i.e. a 367×602 portrait whose top-left sits
// at (137,323); the body layer is the full silhouette, the eyes/brow/mouth are
// small overlays on the head. The old code just shoved that 600px-tall portrait
// up-left with negative offsets so a slice peeked into the panel, which is why
// it never "fit".
//
// Instead: a fixed STAGE sized to the SCALED bounding box, overflow-hidden, with
// an inner CANVAS that scales the whole group and translates the bbox's
// top-left to the stage origin. The internal part alignment is untouched (one
// uniform transform), and the stage has a real height so the panel contains it.
const ORIGIN_X = 137
const ORIGIN_Y = 323
const FRAME_W = 367
const FRAME_H = 602
// The canvas box must span the parts' coordinate space (so absolutely-positioned
// imgs get a real containing block and aren't crushed by a global max-width).
const CANVAS_W = ORIGIN_X + FRAME_W // 504
const CANVAS_H = ORIGIN_Y + FRAME_H // 925
// Scale the ~602px portrait down to a panel-friendly height (~331px).
const SCALE = 0.55
const stageW = Math.round(FRAME_W * SCALE)
const stageH = Math.round(FRAME_H * SCALE)

const canvasStyle = {
  width: `${CANVAS_W}px`,
  height: `${CANVAS_H}px`,
  transform: `scale(${SCALE}) translate(-${ORIGIN_X}px, -${ORIGIN_Y}px)`,
  transformOrigin: 'top left'
}

const loliData = ref({
  loliBodyLeft: '',
  loliBodyTop: '',
  loliEyeLeft: '',
  loliEyeTop: '',
  loliBrowLeft: '',
  loliBrowTop: '',
  loliMouthLeft: '',
  loliMouthTop: '',
  loliFaceLeft: '',
  loliFaceTop: '',
  body: '',
  eye: '',
  brow: '',
  mouth: '',
  face: ''
})

// The 5 webp layers are fetched + blob-encoded on demand; the FIRST load is slow
// (cold cache), during which the stage was blank — so show a KunLoading until it
// resolves. Clicking re-rolls a fresh random character through the same loader.
const isLoading = ref(true)

const reroll = async () => {
  isLoading.value = true
  try {
    loliData.value = await getLoli()
  } finally {
    isLoading.value = false
  }
}

onMounted(reroll)
</script>

<template>
  <div class="hidden shrink-0 sm:block">
    <div
      v-if="isLoading"
      class="flex items-center justify-center"
      :style="{ width: `${stageW}px`, height: `${stageH}px` }"
    >
      <KunLoading />
    </div>

    <KunTooltip v-else-if="loliData.body" text="点击换一个孩子" position="left">
      <div
        class="relative cursor-pointer overflow-hidden"
        :style="{ width: `${stageW}px`, height: `${stageH}px` }"
        @click="reroll"
      >
        <div class="absolute top-0 left-0" :style="canvasStyle">
          <img
            class="absolute max-w-none"
            :src="loliData.body"
            alt="ren"
            :style="{ left: loliData.loliBodyLeft, top: loliData.loliBodyTop }"
          />
          <img
            class="absolute max-w-none"
            :src="loliData.eye"
            alt="ren"
            :style="{ left: loliData.loliEyeLeft, top: loliData.loliEyeTop }"
          />
          <img
            class="absolute max-w-none"
            :src="loliData.brow"
            alt="ren"
            :style="{ left: loliData.loliBrowLeft, top: loliData.loliBrowTop }"
          />
          <img
            class="absolute max-w-none"
            :src="loliData.mouth"
            alt="ren"
            :style="{ left: loliData.loliMouthLeft, top: loliData.loliMouthTop }"
          />
          <img
            class="absolute max-w-none"
            :src="loliData.face"
            alt="ren"
            :style="{ left: loliData.loliFaceLeft, top: loliData.loliFaceTop }"
          />
        </div>
      </div>
    </KunTooltip>
  </div>
</template>
