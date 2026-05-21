<script setup lang="ts">
// Shared field-by-field diff renderer for wiki galgame snapshots
// (snake_case full-state objects). Fed by both:
//   - PR review   : old = base-revision snapshot, new = pr.snapshot
//   - revision diff: GET /revisions/:rev/diff → { changed_keys, old, new }
//
// K-PR4: classify each changed field by value-shape and pick the right
// renderer. Three regimes:
//
//   1. SCALAR             — string / number / boolean.
//      Render: old → new with character-level LCS (useDiff) for short
//      strings; large bodies (intro_*) get the same LCS but in a
//      pre-wrap block so multi-line diffs read naturally.
//
//   2. ARRAY OF OBJECTS   — covers / screenshots / links.
//      Render: arrayDiffByKey ➕added / ➖removed / ✏️changed{fields}.
//      Keys are field-specific (image_hash for cover/screenshot, name+
//      link composite for links) so the renderer doesn't pretend that
//      a sort_order shuffle is a wholesale replace.
//
//   3. ARRAY OF SCALARS   — tag_ids / official_ids / engine_ids / aliases.
//      Render: arrayDiffScalar ➕added / ➖removed pills.
//
// Why this is worth it: presence-replace semantics on the wire mean any
// nontrivial relation edit returns the WHOLE new array; the old "JSON
// stringify both sides + LCS" renderer made even a one-row change look
// like a wall of garbled JSON. Structural diff turns it into a 3-line
// summary the reviewer can act on.
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

// Per-field key/compare config for ARRAY OF OBJECTS regime. Field name
// → identity & relevant scalar fields. Anything else falls back to
// JSON.stringify-LCS so an unknown new wire field still renders
// (degraded, but not broken).
//
// `cdn_url` is intentionally excluded from compareFields — it's a
// server-injected preview URL derived from image_hash; if the hash is
// the same, cdn_url is the same. Including it would inflate noise.
const arrayObjectConfig: Record<
  string,
  { keyFn: (item: Record<string, unknown>) => string; compareFields?: string[] }
> = {
  covers: {
    keyFn: (c) => String(c['image_hash'] ?? ''),
    compareFields: ['sort_order', 'sexual', 'violence']
  },
  screenshots: {
    keyFn: (s) => String(s['image_hash'] ?? ''),
    compareFields: ['sort_order', 'caption', 'sexual', 'violence']
  },
  links: {
    // composite key: same (name,link) pair = same link; reordering
    // does not show as edit.
    keyFn: (l) => `${l['name']}|||${l['link']}`,
    compareFields: []
  }
}

// Field name → true when the value is expected to be an array of
// scalars. Anything else falls into the ARRAY OF OBJECTS or SCALAR
// regimes based on runtime shape.
const SCALAR_ARRAY_FIELDS = new Set([
  'tag_ids',
  'official_ids',
  'engine_ids',
  'aliases'
])

// LARGE_TEXT_FIELDS render in a pre-wrap block so multi-line intro/
// description diffs don't get squeezed onto one line.
const LARGE_TEXT_FIELDS = new Set([
  'intro_en_us',
  'intro_ja_jp',
  'intro_zh_cn',
  'intro_zh_tw',
  'description'
])

type RowKind = 'scalar' | 'array_obj' | 'array_scalar'

interface Row {
  key: string
  label: string
  kind: RowKind
  scalar?: { html: string; preWrap: boolean }
  arrayObj?: {
    added: Record<string, unknown>[]
    removed: Record<string, unknown>[]
    changed: {
      key: string
      old: Record<string, unknown>
      new: Record<string, unknown>
      fields: string[]
    }[]
  }
  arrayScalar?: { added: (string | number)[]; removed: (string | number)[] }
}

const renderScalar = (
  key: string,
  oldVal: unknown,
  newVal: unknown
): Row['scalar'] => {
  const o = oldVal === undefined || oldVal === null ? '' : String(oldVal)
  const n = newVal === undefined || newVal === null ? '' : String(newVal)
  let html: string
  if (!o && n) html = `<b>${n}</b>`
  else if (o && !n) html = `<strong>${o}</strong>`
  // useDiff convention: <b> = added, <strong> = removed
  else html = useDiff(n, o)
  return {
    html: DOMPurify.sanitize(html),
    preWrap: LARGE_TEXT_FIELDS.has(key)
  }
}

const rows = computed<Row[]>(() => {
  const out: Row[] = []
  for (const key of Object.keys(props.changedKeys)) {
    if (!props.changedKeys[key]) continue
    const o = props.oldSnap?.[key]
    const n = props.newSnap?.[key]

    // ARRAY OF OBJECTS
    if (arrayObjectConfig[key]) {
      const cfg = arrayObjectConfig[key]!
      const oArr = Array.isArray(o) ? (o as Record<string, unknown>[]) : []
      const nArr = Array.isArray(n) ? (n as Record<string, unknown>[]) : []
      const d = arrayDiffByKey(
        oArr,
        nArr,
        cfg.keyFn,
        cfg.compareFields as never
      )
      // Skip emitting a row when arrays only differ in order (no add /
      // remove / changed): keeps the diff focused on actual edits.
      if (d.added.length || d.removed.length || d.changed.length) {
        out.push({
          key,
          label: label(key),
          kind: 'array_obj',
          arrayObj: d
        })
      }
      continue
    }

    // ARRAY OF SCALARS
    if (SCALAR_ARRAY_FIELDS.has(key)) {
      const oArr = Array.isArray(o) ? (o as (string | number)[]) : []
      const nArr = Array.isArray(n) ? (n as (string | number)[]) : []
      const d = arrayDiffScalar(oArr, nArr)
      if (d.added.length || d.removed.length) {
        out.push({
          key,
          label: label(key),
          kind: 'array_scalar',
          arrayScalar: d
        })
      }
      continue
    }

    // RUNTIME FALLBACK: a wire field we don't classify statically (e.g.
    // a future new key). If both sides are arrays, use the appropriate
    // shape-based renderer. Otherwise scalar LCS.
    if (Array.isArray(o) || Array.isArray(n)) {
      const oArr = Array.isArray(o) ? o : []
      const nArr = Array.isArray(n) ? n : []
      const allScalars = [...oArr, ...nArr].every(
        (v) => typeof v !== 'object' || v === null
      )
      if (allScalars) {
        const d = arrayDiffScalar(
          oArr as (string | number)[],
          nArr as (string | number)[]
        )
        if (d.added.length || d.removed.length) {
          out.push({
            key,
            label: label(key),
            kind: 'array_scalar',
            arrayScalar: d
          })
        }
        continue
      }
      // unknown array-of-objects → degraded JSON-LCS scalar render
    }

    const scalar = renderScalar(key, o, n)
    if (scalar.html.length > 0) {
      out.push({ key, label: label(key), kind: 'scalar', scalar })
    }
  }
  return out
})

// Compact JSON for an array-of-objects row body (used to show the
// added / removed entries — small objects, single-line).
const inlineRow = (item: Record<string, unknown>): string => {
  return JSON.stringify(item)
}
</script>

<template>
  <div class="kun-diff space-y-4">
    <KunNull v-if="!rows.length" description="无字段变化" />

    <div v-for="row in rows" :key="row.key" class="space-y-1">
      <p class="text-default-500 text-sm font-medium">{{ row.label }}</p>

      <!-- SCALAR -->
      <div
        v-if="row.kind === 'scalar' && row.scalar"
        class="bg-default-100 break-word rounded-md px-3 py-2 text-sm"
        :class="{ 'whitespace-pre-wrap': row.scalar.preWrap }"
        v-html="row.scalar.html"
      />

      <!-- ARRAY OF SCALARS (tag_ids / aliases / ...) -->
      <div
        v-else-if="row.kind === 'array_scalar' && row.arrayScalar"
        class="bg-default-100 flex flex-wrap gap-2 rounded-md px-3 py-2 text-sm"
      >
        <KunChip
          v-for="v in row.arrayScalar.added"
          :key="`+${v}`"
          color="success"
          size="sm"
          variant="flat"
        >
          ➕ {{ v }}
        </KunChip>
        <KunChip
          v-for="v in row.arrayScalar.removed"
          :key="`-${v}`"
          color="danger"
          size="sm"
          variant="flat"
        >
          ➖ {{ v }}
        </KunChip>
      </div>

      <!-- ARRAY OF OBJECTS (covers / screenshots / links) -->
      <div
        v-else-if="row.kind === 'array_obj' && row.arrayObj"
        class="bg-default-100 space-y-2 rounded-md px-3 py-2 text-sm"
      >
        <div
          v-for="(item, i) in row.arrayObj.added"
          :key="`add-${i}`"
          class="flex items-start gap-2"
        >
          <KunChip color="success" size="sm" variant="flat">➕ 新增</KunChip>
          <code class="break-all text-xs">{{ inlineRow(item) }}</code>
        </div>
        <div
          v-for="(item, i) in row.arrayObj.removed"
          :key="`rm-${i}`"
          class="flex items-start gap-2"
        >
          <KunChip color="danger" size="sm" variant="flat">➖ 删除</KunChip>
          <code class="break-all text-xs">{{ inlineRow(item) }}</code>
        </div>
        <div
          v-for="(ch, i) in row.arrayObj.changed"
          :key="`ch-${i}`"
          class="space-y-1"
        >
          <div class="flex items-center gap-2">
            <KunChip color="warning" size="sm" variant="flat">
              ✏️ 修改
            </KunChip>
            <span class="text-default-400 text-xs">
              {{ ch.fields.join(' / ') }}
            </span>
          </div>
          <div class="text-default-500 ml-2 text-xs">
            旧: <code class="break-all">{{ inlineRow(ch.old) }}</code>
          </div>
          <div class="text-default-700 ml-2 text-xs">
            新: <code class="break-all">{{ inlineRow(ch.new) }}</code>
          </div>
        </div>
      </div>
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
