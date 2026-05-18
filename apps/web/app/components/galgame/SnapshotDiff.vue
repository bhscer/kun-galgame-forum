<script setup lang="ts">
// Shared field-by-field diff renderer for wiki galgame snapshots
// (snake_case full-state objects). Fed by both:
//   - PR review  : old = base-revision snapshot, new = pr.snapshot
//   - revision diff: GET /revisions/:rev/diff → { changed_keys, old, new }
//
// Replaces the old client-side compare.ts (which diffed a now-defunct
// oldData/newData shape — the wiki PR/revision API returns
// {changed_keys, old/snapshot} instead, see docs 02-revisions-and-prs).
import DOMPurify from 'isomorphic-dompurify'
import { KUN_GALGAME_RESOURCE_PULL_REQUEST_I18N_FIELD_MAP } from '~/constants/galgame'

const props = defineProps<{
  changedKeys: Record<string, boolean>
  oldSnap: Record<string, unknown>
  newSnap: Record<string, unknown>
}>()

const label = (key: string): string => {
  const m = KUN_GALGAME_RESOURCE_PULL_REQUEST_I18N_FIELD_MAP[key]
  return typeof m === 'string' ? m : key
}

// Stringify a snapshot value for display. Scalars as-is; arrays/objects
// (aliases / tag_ids / links …) as compact JSON so a change is visible
// even if not pretty.
const toStr = (v: unknown): string => {
  if (v === undefined || v === null) return ''
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

const rows = computed(() => {
  return Object.keys(props.changedKeys)
    .filter((k) => props.changedKeys[k])
    .map((key) => {
      const oldStr = toStr(props.oldSnap?.[key])
      const newStr = toStr(props.newSnap?.[key])
      let html: string
      if (!oldStr && newStr) {
        html = `<b>${newStr}</b>`
      } else if (oldStr && !newStr) {
        html = `<strong>${oldStr}</strong>`
      } else {
        // useDiff: <b> = added, <strong> = removed
        html = useDiff(newStr, oldStr)
      }
      return { key, label: label(key), html: DOMPurify.sanitize(html) }
    })
    .filter((r) => r.html.length > 0)
})
</script>

<template>
  <div class="kun-diff space-y-3">
    <KunNull v-if="!rows.length" description="无字段变化" />
    <div v-for="row in rows" :key="row.key">
      <p class="text-default-500 mb-1 text-sm font-medium">{{ row.label }}</p>
      <div
        class="bg-default-100 break-word rounded-md px-3 py-2 text-sm"
        v-html="row.html"
      />
    </div>
  </div>
</template>

<style scoped>
.kun-diff {
  :deep(b) {
    color: var(--color-success);
    background-color: color-mix(in oklab, var(--color-success) 20%, transparent);
  }
  :deep(strong) {
    color: var(--color-danger);
    background-color: color-mix(in oklab, var(--color-danger) 20%, transparent);
    text-decoration: line-through;
  }
}
</style>
