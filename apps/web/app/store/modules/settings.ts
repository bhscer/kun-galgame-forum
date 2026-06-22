import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ENABLE_KUN_VISUAL_NOVEL_FORUM_WINTER_THEME,
  KUN_VISUAL_NOVEL_FORUM_WINTER_THEME_BACKGROUND
} from '~/config/theme'
import type { KUNGalgameSettingsStore } from '../types/settings'

const SETTINGS_CUSTOM_BACKGROUND_IMAGE_NAME: string = 'kun-galgame-custom-bg'
const SETTINGS_PUBLISH_Banner_IMAGE_NAME: string = 'kun-galgame-publish-banner'
const SETTINGS_DEFAULT_FONT_FAMILY: string = 'system-ui'

export const usePersistSettingsStore = defineStore(
  'KUNGalgameSettings',
  () => {
    const showKUNGalgamePageTransparency =
      ref<KUNGalgameSettingsStore['showKUNGalgamePageTransparency']>(80)
    const showKUNGalgameFontStyle = ref<
      KUNGalgameSettingsStore['showKUNGalgameFontStyle']
    >(SETTINGS_DEFAULT_FONT_FAMILY)
    const showKUNGalgameContentLimit =
      ref<KUNGalgameSettingsStore['showKUNGalgameContentLimit']>('sfw')
    const showKUNGalgameBackground =
      ref<KUNGalgameSettingsStore['showKUNGalgameBackground']>(0)
    const showKUNGalgameBackgroundBlur =
      ref<KUNGalgameSettingsStore['showKUNGalgameBackgroundBlur']>(0)
    const showKUNGalgameBackgroundBrightness =
      ref<KUNGalgameSettingsStore['showKUNGalgameBackgroundBrightness']>(100)
    const showKUNGalgameBackLoli =
      ref<KUNGalgameSettingsStore['showKUNGalgameBackLoli']>(false)
    // Global "显示没有下载资源的 Galgame" — off by default; hides resource-less
    // galgames across all local galgame lists. See KUNGalgameSettingsStore.
    const showKUNGalgameNoResource =
      ref<KUNGalgameSettingsStore['showKUNGalgameNoResource']>(false)
    // Global corner-radius level (直角/小/中/大). One knob rounds the WHOLE UI:
    // both the forum's own rounded-* and KunUI's rounded-kun-* derive from the
    // --kun-radius-scale CSS multiplier this sets (see styles/tailwindcss.css).
    // 'md' = stock look.
    const showKUNGalgameRounded =
      ref<KUNGalgameSettingsStore['showKUNGalgameRounded']>('md')
    // Per-rating galgame gallery filters — the rating levels (1/2/3) the viewer
    // opted to reveal. Default [] = only unrated shows. 色情 is additionally
    // gated by the global NSFW mode; 暴力 is an independent warned opt-in. See
    // KUNGalgameSettingsStore + components/galgame/GalleryFilter.vue.
    const showKUNGalgameGallerySexualLevels =
      ref<KUNGalgameSettingsStore['showKUNGalgameGallerySexualLevels']>([])
    const showKUNGalgameGalleryViolenceLevels =
      ref<KUNGalgameSettingsStore['showKUNGalgameGalleryViolenceLevels']>([])

    const setKUNGalgameFontStyle = (font: string) => {
      showKUNGalgameFontStyle.value = font
      document.documentElement.style.setProperty('--font-family', font)
    }

    const setKUNGalgameTransparency = (trans: number) => {
      showKUNGalgamePageTransparency.value = trans
      const opacity = `${trans / 100}`
      // Page background + default-100 glass.
      document.documentElement.style.setProperty('--kun-global-opacity', opacity)
      // Raised surfaces (cards / inputs / modals / dropdowns): KunUI 1.8 split
      // these onto --kun-surface-opacity (default 1 = opaque). Drive it from the
      // same slider so surfaces stay translucent over a background image. Blur is
      // a separate opt-in knob (--kun-background-blur, default 0) — left untouched.
      document.documentElement.style.setProperty('--kun-surface-opacity', opacity)
    }

    const setKUNGalgameBackgroundBlur = (blur: number) => {
      showKUNGalgameBackgroundBlur.value = blur
      document.documentElement.style.setProperty(
        '--kun-background-blur',
        `${blur}px`
      )
    }

    const setKUNGalgameBackgroundBrightness = (brightness: number) => {
      showKUNGalgameBackgroundBrightness.value = brightness
      document.documentElement.style.setProperty(
        '--kun-background-brightness',
        `${brightness}%`
      )
    }

    // Radius multiplier per level. md (1) keeps every radius at its stock
    // value; the rest scale the whole hierarchy proportionally. Both the
    // forum's --radius-* and KunUI's --radius-kun-* derive from this one
    // multiplier (styles/tailwindcss.css), so it rounds everything at once.
    const ROUNDED_SCALE: Record<
      KUNGalgameSettingsStore['showKUNGalgameRounded'],
      number
    > = { none: 0, sm: 0.5, md: 1, lg: 1.5 }

    const setKUNGalgameRounded = (
      level: KUNGalgameSettingsStore['showKUNGalgameRounded']
    ) => {
      showKUNGalgameRounded.value = level
      document.documentElement.style.setProperty(
        '--kun-radius-scale',
        `${ROUNDED_SCALE[level]}`
      )
    }

    const setSystemBackground = async (index: number) => {
      showKUNGalgameBackground.value = index
      await deleteImage(SETTINGS_CUSTOM_BACKGROUND_IMAGE_NAME)
    }

    const setCustomBackground = async (file: File) => {
      await saveImage(file, SETTINGS_CUSTOM_BACKGROUND_IMAGE_NAME)
      showKUNGalgameBackground.value = -1
    }

    const getCurrentBackground = async () => {
      const backgroundImageBlobData = await getImage(
        SETTINGS_CUSTOM_BACKGROUND_IMAGE_NAME
      )
      if (showKUNGalgameBackground.value === 0) {
        return ENABLE_KUN_VISUAL_NOVEL_FORUM_WINTER_THEME
          ? KUN_VISUAL_NOVEL_FORUM_WINTER_THEME_BACKGROUND
          : ''
      }

      if (showKUNGalgameBackground.value === -1 && backgroundImageBlobData) {
        return URL.createObjectURL(backgroundImageBlobData)
      }

      return `/bg/bg${showKUNGalgameBackground.value}.webp`
    }

    const setKUNGalgameSettingsRecover = async () => {
      kungalgameStoreReset()
      await deleteImage(SETTINGS_CUSTOM_BACKGROUND_IMAGE_NAME)
      await deleteImage(SETTINGS_PUBLISH_Banner_IMAGE_NAME)
    }

    return {
      showKUNGalgamePageTransparency,
      showKUNGalgameFontStyle,
      showKUNGalgameContentLimit,
      showKUNGalgameBackground,
      showKUNGalgameBackgroundBlur,
      showKUNGalgameBackgroundBrightness,
      showKUNGalgameBackLoli,
      showKUNGalgameNoResource,
      showKUNGalgameRounded,
      showKUNGalgameGallerySexualLevels,
      showKUNGalgameGalleryViolenceLevels,
      setKUNGalgameFontStyle,
      setKUNGalgameTransparency,
      setKUNGalgameBackgroundBlur,
      setKUNGalgameBackgroundBrightness,
      setKUNGalgameRounded,
      setSystemBackground,
      setCustomBackground,
      getCurrentBackground,
      setKUNGalgameSettingsRecover
    }
  },
  {
    persist: true
  }
)
