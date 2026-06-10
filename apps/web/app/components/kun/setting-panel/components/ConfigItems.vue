<script setup lang="ts">
const showItemIndex = ref(1)

const {
  showKUNGalgamePageTransparency,
  showKUNGalgameBackgroundBlur,
  showKUNGalgameBackgroundBrightness,
  showKUNGalgameRounded
} = storeToRefs(usePersistSettingsStore())

const roundedOptions = [
  { value: 'none', label: '直角' },
  { value: 'sm', label: '小' },
  { value: 'md', label: '中' },
  { value: 'lg', label: '大' }
] as const

watch(
  () => showKUNGalgamePageTransparency.value,
  debounce(() => {
    usePersistSettingsStore().setKUNGalgameTransparency(
      showKUNGalgamePageTransparency.value
    )
  }, 300)
)

watch(
  () => showKUNGalgameBackgroundBlur.value,
  debounce(() => {
    usePersistSettingsStore().setKUNGalgameBackgroundBlur(
      showKUNGalgameBackgroundBlur.value
    )
  }, 300)
)

watch(
  () => showKUNGalgameBackgroundBrightness.value,
  debounce(() => {
    usePersistSettingsStore().setKUNGalgameBackgroundBrightness(
      showKUNGalgameBackgroundBrightness.value
    )
  }, 300)
)

</script>

<template>
  <div class="flex flex-col space-y-2">
    <div class="flex items-center justify-between">
      可配置选项

      <div class="flex items-center">
        <span
          :class="
            cn(
              'flex rounded-lg p-2 transition-colors',
              showItemIndex === 1 ? 'bg-primary-50 text-primary' : ''
            )
          "
          @click="showItemIndex = 1"
        >
          <KunIcon class="text-inherit" name="mdi:circle-transparent" />
        </span>
        <span
          :class="
            cn(
              'flex rounded-lg p-2 transition-colors',
              showItemIndex === 2 ? 'bg-primary-50 text-primary' : ''
            )
          "
          @click="showItemIndex = 2"
        >
          <KunIcon class="text-inherit" name="tabler:blur" />
        </span>
        <span
          :class="
            cn(
              'flex rounded-lg p-2 transition-colors',
              showItemIndex === 3 ? 'bg-primary-50 text-primary' : ''
            )
          "
          @click="showItemIndex = 3"
        >
          <KunIcon class="text-inherit" name="lucide:lightbulb" />
        </span>
        <span
          :class="
            cn(
              'flex rounded-lg p-2 transition-colors',
              showItemIndex === 4 ? 'bg-primary-50 text-primary' : ''
            )
          "
          @click="showItemIndex = 4"
        >
          <KunIcon class="text-inherit" name="tabler:border-radius" />
        </span>
      </div>
    </div>

    <TransitionGroup name="item" tag="div">
      <div class="w-full space-y-2" v-if="showItemIndex === 1">
        <div class="flex justify-between text-sm">
          <span>页面透明度</span>
          <span>{{ showKUNGalgamePageTransparency }}%</span>
        </div>

        <div class="flex items-center">
          <span>10%</span>

          <KunSlider
            class="mx-4 w-full"
            :min="10"
            :max="90"
            :step="1"
            v-model="showKUNGalgamePageTransparency"
          />

          <span>90%</span>
        </div>
      </div>

      <div class="w-full space-y-2" v-if="showItemIndex === 2">
        <div class="flex justify-between text-sm">
          <span>页面模糊度</span>
          <span>{{ showKUNGalgameBackgroundBlur }}px</span>
        </div>

        <div class="flex w-full items-center">
          <span>0px</span>
          <KunSlider
            class="mx-4 w-full"
            :min="0"
            :max="32"
            :step="1"
            v-model="showKUNGalgameBackgroundBlur"
          />
          <span>32px</span>
        </div>
      </div>

      <div class="w-full space-y-2" v-if="showItemIndex === 3">
        <div class="flex justify-between text-sm">
          <span>背景亮度</span>
          <span>{{ showKUNGalgameBackgroundBrightness }}%</span>
        </div>

        <div class="flex w-full items-center">
          <span>10%</span>
          <KunSlider
            class="mx-4 w-full"
            :min="10"
            :max="100"
            :step="1"
            v-model="showKUNGalgameBackgroundBrightness"
          />
          <span>100%</span>
        </div>
      </div>

      <div class="w-full space-y-2" v-if="showItemIndex === 4">
        <div class="flex justify-between text-sm">
          <span>全局圆角</span>
        </div>

        <div class="grid grid-cols-4 gap-2">
          <KunButton
            v-for="opt in roundedOptions"
            :key="opt.value"
            size="sm"
            :variant="showKUNGalgameRounded === opt.value ? 'solid' : 'flat'"
            :color="showKUNGalgameRounded === opt.value ? 'primary' : 'default'"
            @click="usePersistSettingsStore().setKUNGalgameRounded(opt.value)"
          >
            {{ opt.label }}
          </KunButton>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style  scoped>
.item-move,
.item-enter-active,
.item-leave-active {
  transition: all 0.5s ease;
}

.item-enter-from,
.item-leave-to {
  opacity: 0;
  transform: translateY(77px);
}

.item-leave-active {
  position: absolute;
}
</style>
