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
  // Release-year bounds, wiki §17. Year-granularity ('' | 'YYYY');
  // single-year pick keeps from === to. Month is no longer encoded here
  // — it's the orthogonal `releasedMonths` set below.
  releasedFrom: string
  releasedTo: string
  // Discontinuous month set, wiki §17.10: csv of 1–12 (e.g. '3,7').
  // AND-combined with the year range; works with OR without a year
  // ('历年三月' = releasedMonths='3' + no year). Spread into the query.
  releasedMonths: string
}
