<script setup lang="ts" generic="T extends { id: number; name: string }">
// Reusable "search-and-pick-many" selector for wiki entities (tags /
// officials / engines) on the PR edit form. The three concrete pickers
// (TagSelector / OfficialSelector / EngineSelector) only differ in their
// data source, so they inject a `search` fn and bind their own store
// array; all the dropdown / debounce / dedupe / badge UX lives here.
//
// Mirrors the established search-dropdown UX from
// components/galgame/tag/Container.vue (relative wrapper + absolute
// dropdown, @mousedown.prevent so the click lands before blur, 120ms
// blur close). Selected entities render as removable KunChip.
//
// The model holds FULL entity objects (not ids): callers store these in
// the temp PR store so names render without a second lookup, and
// Footer.vue derives the *_ids[] wire arrays at submit.
import { watchDebounced } from '@vueuse/core'

const props = defineProps<{
  title: string
  description?: string
  placeholder: string
  // Resolve a query string to candidate entities. Async covers both the
  // network pickers (tag/official) and the local-filter one (engine).
  search: (q: string) => Promise<T[]>
  // When true, a "create「query」" row shows if no existing entity name
  // matches — the doc-mandated "选已有 + 没有就新建" interaction
  // (00-handbook §15 / 04-taxonomy). The parent handles `create` (opens
  // a per-entity create modal) then calls addCreated() with the result.
  creatable?: boolean
}>()

const emit = defineEmits<{ create: [query: string] }>()

const selected = defineModel<T[]>({ required: true })

const query = ref('')
const results = ref<T[]>([]) as Ref<T[]>
const focused = ref(false)
const searching = ref(false)

const run = async () => {
  const q = query.value.trim()
  if (!q) {
    results.value = []
    return
  }
  searching.value = true
  try {
    results.value = await props.search(q)
  } finally {
    searching.value = false
  }
}

watchDebounced(() => query.value, run, { debounce: 500, maxWait: 1000 })

// Offer "create" only when the trimmed query has no case-insensitive
// exact name match among the current results (an inexact/substring hit
// shouldn't block creating a distinct new entity).
const canCreate = computed(() => {
  const q = query.value.trim()
  if (!props.creatable || !q) return false
  const lower = q.toLowerCase()
  return !results.value.some((e) => e.name.toLowerCase() === lower)
})

const isDropdownOpen = computed(
  () =>
    focused.value &&
    !!query.value.trim() &&
    (!!results.value.length || canCreate.value)
)

const onBlur = () => {
  setTimeout(() => {
    focused.value = false
  }, 120)
}

const add = (item: T) => {
  if (!selected.value.some((e) => e.id === item.id)) {
    selected.value = [...selected.value, item]
  }
  query.value = ''
  results.value = []
}

const remove = (id: number) => {
  selected.value = selected.value.filter((e) => e.id !== id)
}

// Parent calls this after its create modal resolves: dedupe + select +
// reset, identical to picking an existing one.
defineExpose({ addCreated: add })
</script>

<template>
  <div class="relative space-y-2">
    <h2 class="text-xl">{{ title }}</h2>
    <p v-if="description" class="text-default-500 text-sm">
      {{ description }}
    </p>

    <div v-if="selected.length" class="flex flex-wrap gap-2">
      <KunChip
        v-for="e in selected"
        :key="e.id"
        color="primary"
        size="md"
        class="flex items-center gap-1"
      >
        {{ e.name }}
        <button
          type="button"
          class="text-primary ml-1 focus:outline-none"
          @click="remove(e.id)"
        >
          ×
        </button>
      </KunChip>
    </div>

    <div class="relative">
      <KunInput
        v-model="query"
        type="text"
        :placeholder="placeholder"
        @focus="focused = true"
        @blur="onBlur"
      />

      <div
        v-if="isDropdownOpen"
        class="border-default-200 bg-content1 absolute top-full right-0 left-0 z-50 mt-2 rounded-lg border shadow-md"
      >
        <KunScrollShadow
          axis="vertical"
          shadow-size="3rem"
          class-name="max-h-72"
        >
          <div
            v-for="item in results"
            :key="item.id"
            class="hover:bg-default-100 flex cursor-pointer items-center justify-between px-3 py-2"
            @mousedown.prevent="add(item)"
          >
            <span class="truncate">{{ item.name }}</span>
            <KunIcon
              v-if="selected.some((s) => s.id === item.id)"
              name="lucide:check"
              class="text-success shrink-0"
            />
          </div>

          <div
            v-if="canCreate"
            class="text-primary border-default-200 hover:bg-default-100 flex cursor-pointer items-center gap-2 border-t px-3 py-2"
            @mousedown.prevent="emit('create', query.trim())"
          >
            <KunIcon name="lucide:plus" class="shrink-0" />
            <span class="truncate">新建「{{ query.trim() }}」</span>
          </div>
        </KunScrollShadow>
      </div>
    </div>
  </div>
</template>
