<script setup lang="ts">
import { getLoli } from './getLoli'

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

// The 差分图 is 5 webp layers fetched + blob-encoded on demand. The FIRST load
// is slow (cold cache, several images downloaded in parallel), during which the
// panel showed nothing — so surface a KunLoading until it resolves. Clicking
// re-rolls a fresh random character (usually warm-cached, but the loader covers
// it too).
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
  <div class="hidden w-80 shrink-0 sm:block">
    <div v-if="isLoading" class="flex min-h-80 items-center justify-center">
      <KunLoading />
    </div>

    <div
      v-else-if="loliData.body"
      class="relative min-w-80"
      @click="reroll"
    >
      <div class="absolute -top-80 -left-32 size-full shrink-0">
        <img
          class="absolute shrink-0"
          :src="loliData.body"
          alt="ren"
          :style="{ left: loliData.loliBodyLeft, top: loliData.loliBodyTop }"
        />
      </div>
      <div class="absolute -top-80 -left-32 size-full shrink-0">
        <img
          class="absolute"
          :src="loliData.eye"
          alt="ren"
          :style="{ left: loliData.loliEyeLeft, top: loliData.loliEyeTop }"
        />
      </div>
      <div class="absolute -top-80 -left-32 size-full shrink-0">
        <img
          class="absolute"
          :src="loliData.brow"
          alt="ren"
          :style="{ left: loliData.loliBrowLeft, top: loliData.loliBrowTop }"
        />
      </div>
      <div class="absolute -top-80 -left-32 size-full shrink-0">
        <img
          class="absolute"
          :src="loliData.mouth"
          alt="ren"
          :style="{ left: loliData.loliMouthLeft, top: loliData.loliMouthTop }"
        />
      </div>
      <div class="absolute -top-80 -left-32 size-full shrink-0">
        <img
          class="absolute"
          :src="loliData.face"
          alt="ren"
          :style="{ left: loliData.loliFaceLeft, top: loliData.loliFaceTop }"
        />
      </div>
    </div>
  </div>
</template>
