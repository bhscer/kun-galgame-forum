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
}

export interface TempSettingStore {
  showKUNGalgameHamburger: boolean
  showKUNGalgamePanel: boolean
  showKUNGalgameUserPanel: boolean

  showKUNGalgameMessageBox: boolean
  showKUNGalgameMoemoepointLog: boolean
  messageStatus: MessageStatus
}
