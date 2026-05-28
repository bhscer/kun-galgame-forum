<script setup lang="ts">
import {
  KUN_GALGAME_RESOURCE_TYPE_MAP,
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP,
  KUN_GALGAME_RESOURCE_SORT_FIELD_MAP
} from '~/constants/galgame'
import {
  KUN_GALGAME_PROVIDER_LABEL_MAP,
  PROVIDER_KEY_OPTIONS,
  type ProviderKey
} from '~/constants/galgameResource'
import { usePersistKUNGalgameAdvancedFilterStore } from '~/store/modules/galgame'
import type {
  KunGalgameResourceTypeOptions,
  KunGalgameResourceLanguageOptions,
  KunGalgameResourcePlatformOptions
} from '~/constants/galgame'

withDefaults(
  defineProps<{
    isShowAdvanced?: boolean
  }>(),
  { isShowAdvanced: false }
)

const {
  page,
  type,
  language,
  platform,
  sortField,
  sortOrder,
  releasedFrom,
  releasedTo,
  releasedMonths
} = storeToRefs(useTempGalgameStore())

const advStore = usePersistKUNGalgameAdvancedFilterStore()
const { includeProviders, excludeOnlyProviders } = storeToRefs(advStore)
const showAdvanced = ref(false)
const showReleaseFilter = ref(false)

watch(
  () => [
    type.value,
    language.value,
    platform.value,
    sortField.value,
    sortOrder.value,
    releasedFrom.value,
    releasedMonths.value,
    includeProviders.value.join(','),
    excludeOnlyProviders.value.join(',')
  ],
  () => {
    page.value = 1
  }
)

// ── Release year / month filter (wiki §17 + §17.10) ─────────────────
// Two ORTHOGONAL controls, both derived views over the store so the
// selection survives Nav remounts (store is the source of truth):
//   • year range — `releasedFrom`/`releasedTo` (each '' | 'YYYY',
//             independent). Pick the same on both ends for a single
//             year; leave one end '全部/不限' for an open-ended range
//             ("2020 及以后" / "2024 及以前").
//   • months — multi-select set (wiki §17.10) → releasedMonths csv
//             ('' | '3' | '3,7'). AND-combined with the year range, and
//             works WITHOUT a year ('历年三月' = months only, no year).
// Mirrors the 网盘筛选 panel's toggle-button-+-chip-rows UX.
const KUN_RELEASE_EARLIEST_YEAR = 1980
const yearOptions = [
  { value: '', label: '不限' },
  ...Array.from(
    { length: new Date().getFullYear() - KUN_RELEASE_EARLIEST_YEAR + 1 },
    (_, i) => {
      const y = String(new Date().getFullYear() - i)
      return { value: y, label: `${y}` }
    }
  )
]
const monthOptions = Array.from({ length: 12 }, (_, i) => ({
  value: i + 1,
  label: `${i + 1} 月`
}))

// Year-range setters with clamping so the lower bound never exceeds the
// upper (picking a `from` past the current `to` drags `to` along, and
// vice versa) — an inverted range would silently return nothing.
const applyFromYear = (year: string) => {
  releasedFrom.value = year
  if (year && releasedTo.value && Number(releasedTo.value) < Number(year)) {
    releasedTo.value = year
  }
}
const applyToYear = (year: string) => {
  releasedTo.value = year
  if (year && releasedFrom.value && Number(releasedFrom.value) > Number(year)) {
    releasedFrom.value = year
  }
}

// Month set parsed from the csv store value, for highlight checks.
const selectedMonths = computed(
  () => new Set(releasedMonths.value.split(',').filter(Boolean))
)
const isMonthSelected = (m: number) => selectedMonths.value.has(String(m))

// Toggle a month in/out of the set, re-serialising sorted csv.
const toggleMonth = (m: number) => {
  const set = new Set(selectedMonths.value)
  const key = String(m)
  if (set.has(key)) {
    set.delete(key)
  } else {
    set.add(key)
  }
  releasedMonths.value = [...set]
    .map(Number)
    .sort((a, b) => a - b)
    .join(',')
}

const typeOptions = Object.entries(KUN_GALGAME_RESOURCE_TYPE_MAP)
  .filter(([k]) => k !== 'name')
  .map(([value, label]) => ({ value, label }))

const langOptions = Object.entries(KUN_GALGAME_RESOURCE_LANGUAGE_MAP).map(
  ([value, label]) => ({ value, label })
)

const platformOptions = Object.entries(KUN_GALGAME_RESOURCE_PLATFORM_MAP)
  .filter(([k]) => k !== 'name')
  .map(([value, label]) => ({ value, label }))

const sortOptions = Object.entries(KUN_GALGAME_RESOURCE_SORT_FIELD_MAP).map(
  ([value, label]) => ({
    value: value === 'views' ? 'view' : value,
    label
  })
)
</script>

<template>
  <div class="space-y-1">
    <KunScrollShadow>
      <button
        v-for="opt in typeOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          type === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="type = opt.value as KunGalgameResourceTypeOptions"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in langOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          language === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="language = opt.value as KunGalgameResourceLanguageOptions"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in platformOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          platform === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="platform = opt.value as KunGalgameResourcePlatformOptions"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in sortOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          sortField === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="
          sortField = opt.value as 'time' | 'view' | 'created' | 'release_date'
        "
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <div class="flex items-center gap-1.5">
      <button
        class="cursor-pointer rounded-md p-1 transition-colors"
        :class="
          sortOrder === 'desc'
            ? 'bg-primary/15 text-primary'
            : 'text-default-500 hover:bg-default-100'
        "
        @click="sortOrder = 'desc'"
      >
        <KunIcon name="lucide:arrow-down" />
      </button>
      <button
        class="cursor-pointer rounded-md p-1 transition-colors"
        :class="
          sortOrder === 'asc'
            ? 'bg-primary/15 text-primary'
            : 'text-default-500 hover:bg-default-100'
        "
        @click="sortOrder = 'asc'"
      >
        <KunIcon name="lucide:arrow-up" />
      </button>

      <button
        v-if="isShowAdvanced"
        class="text-default-500 hover:text-primary flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm transition-colors"
        :class="
          (includeProviders.length || excludeOnlyProviders.length) &&
          'text-warning'
        "
        @click="showAdvanced = !showAdvanced"
      >
        <KunIcon name="lucide:filter" class="text-inherit" />
        <span>网盘筛选</span>
      </button>

      <button
        v-if="isShowAdvanced"
        class="text-default-500 hover:text-primary flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm transition-colors"
        :class="(releasedFrom || releasedMonths) && 'text-warning'"
        @click="showReleaseFilter = !showReleaseFilter"
      >
        <KunIcon name="lucide:calendar-days" class="text-inherit" />
        <span>发售日期</span>
      </button>
    </div>

    <div
      v-if="showReleaseFilter"
      class="bg-default-50 space-y-3 rounded-lg border p-3"
    >
      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          起始年份
        </div>
        <KunScrollShadow>
          <button
            v-for="opt in yearOptions"
            :key="opt.value || 'from-all'"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              releasedFrom === opt.value
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="applyFromYear(opt.value)"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          结束年份
        </div>
        <KunScrollShadow>
          <button
            v-for="opt in yearOptions"
            :key="opt.value || 'to-all'"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              releasedTo === opt.value
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="applyToYear(opt.value)"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          发售月份
          <span class="text-default-400 font-normal">(可多选, 含历年)</span>
        </div>
        <KunScrollShadow>
          <button
            v-for="opt in monthOptions"
            :key="opt.value"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              isMonthSelected(opt.value)
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="toggleMonth(opt.value)"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>
    </div>

    <div
      v-if="showAdvanced"
      class="bg-default-50 space-y-3 rounded-lg border p-3"
    >
      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          必须含有以下网盘
        </div>
        <KunScrollShadow>
          <button
            v-for="key in PROVIDER_KEY_OPTIONS"
            :key="key"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              includeProviders.includes(key)
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="advStore.toggleIncludeProvider(key as ProviderKey)"
          >
            {{ KUN_GALGAME_PROVIDER_LABEL_MAP[key as ProviderKey] }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          排除仅含以下网盘
        </div>
        <KunScrollShadow>
          <button
            v-for="key in PROVIDER_KEY_OPTIONS"
            :key="key + '-ex'"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              excludeOnlyProviders.includes(key)
                ? 'bg-danger/15 text-danger font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="advStore.toggleExcludeOnlyProvider(key as ProviderKey)"
          >
            {{ KUN_GALGAME_PROVIDER_LABEL_MAP[key as ProviderKey] }}
          </button>
        </KunScrollShadow>
      </div>
    </div>
  </div>
</template>
