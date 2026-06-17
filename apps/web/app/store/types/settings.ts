import type { MessageStatus } from '~~/shared/types/utils/message'

export interface KUNGalgameSettingsStore {
  showKUNGalgamePageTransparency: number
  showKUNGalgameFontStyle: string
  showKUNGalgameContentLimit: string
  showKUNGalgameBackground: number
  showKUNGalgameBackgroundBlur: number
  showKUNGalgameBackgroundBrightness: number
  showKUNGalgameBackLoli: boolean
  showKUNGalgameSidebarCollapsed: boolean
  // Global "显示没有下载资源的 Galgame" toggle. false (default) hides
  // resource-less galgames across all local galgame lists (browse / ranking /
  // user pages — NOT wiki-proxied entity pages or search). Cookie-persisted
  // (this store) so it's SSR-readable + remembered.
  showKUNGalgameNoResource: boolean
  // Global corner-radius level (直角/小/中/大, default 'md'). Maps to a
  // --kun-radius-scale multiplier that rounds both the forum and KunUI at once.
  showKUNGalgameRounded: 'none' | 'sm' | 'md' | 'lg'
  // Per-rating galgame gallery filters (detail 画廊). Each holds the rating
  // LEVELS (1=轻 / 2=中 / 3=高) the viewer opted to reveal. 色情 is also gated
  // by showKUNGalgameContentLimit (NSFW shows every level); 暴力 is INDEPENDENT
  // of it — NSFW never auto-reveals violence, it's an explicit warned opt-in.
  // Default [] = only unrated (level 0) shows. Persisted so it's remembered.
  showKUNGalgameGallerySexualLevels: number[]
  showKUNGalgameGalleryViolenceLevels: number[]
}

export interface TempSettingStore {
  showKUNGalgameHamburger: boolean
  showKUNGalgamePanel: boolean
  showKUNGalgameUserPanel: boolean

  showKUNGalgameMessageBox: boolean
  showKUNGalgameMoemoepointLog: boolean
  showKUNGalgameLogout: boolean
  messageStatus: MessageStatus
}
