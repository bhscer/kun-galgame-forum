import type {
  KunGalgameResourceTypeOptions,
  KunGalgameResourceLanguageOptions,
  KunGalgameResourcePlatformOptions
} from '~/constants/galgame'

export interface GalgameStoreTemp {
  page: number
  limit: number
  type: KunGalgameResourceTypeOptions
  language: KunGalgameResourceLanguageOptions
  platform: KunGalgameResourcePlatformOptions
  sortField: 'time' | 'created' | 'view' | 'release_date'
  sortOrder: KunOrder
  // Release-date filter bounds, wiki §17 format: '' | 'YYYY' | 'YYYY-MM'.
  // For a single year/month pick from === to; spread straight into the
  // /galgame query as releasedFrom / releasedTo.
  releasedFrom: string
  releasedTo: string
}
