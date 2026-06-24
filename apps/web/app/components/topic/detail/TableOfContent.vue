<script setup lang="ts">
import type { ComponentPublicInstance } from 'vue'
import { useTopicTOC, TOPIC_TOC_SOURCE } from '~/composables/topic/useTopicTOC'
import { scrollToTOCElement } from '../_helper'

// The list is derived from the SSR'd topic data (provided by Detail.vue) so the
// rail renders server-side; the scrollspy inside useTopicTOC stays client-only.
const source = inject(TOPIC_TOC_SOURCE)!
const { headings, activeIds, refreshTOC } = useTopicTOC(source)

defineExpose({ refreshTOC })

type GroupType = 'heading' | 'reply'
const GROUPS: { type: GroupType; title: string }[] = [
  { type: 'heading', title: '文章目录' },
  { type: 'reply', title: '回复列表' }
]

// One highlight per group: an absolutely-positioned overlay covering the
// CONTIGUOUS run of active items, so the whole region reads as a single block
// and its top/height transition smoothly (grow/shrink) as items join or leave —
// rather than each item lighting up on its own.
const liRefs = new Map<string, HTMLElement>()
const wrapperRefs = new Map<GroupType, HTMLElement>()

const setLiRef = (id: string, el: Element | ComponentPublicInstance | null) => {
  if (el instanceof HTMLElement) liRefs.set(id, el)
  else liRefs.delete(id)
}
const setWrapperRef = (
  type: GroupType,
  el: Element | ComponentPublicInstance | null
) => {
  if (el instanceof HTMLElement) wrapperRefs.set(type, el)
  else wrapperRefs.delete(type)
}

const overlays = reactive<
  Record<GroupType, { top: number; height: number; visible: boolean }>
>({
  heading: { top: 0, height: 0, visible: false },
  reply: { top: 0, height: 0, visible: false }
})

const measure = () => {
  for (const { type } of GROUPS) {
    const wrapper = wrapperRefs.get(type)
    const els = headings.value
      .filter((h) => h.type === type && activeIds.value.includes(h.id))
      .map((h) => liRefs.get(h.id))
      .filter((el): el is HTMLElement => !!el)

    if (!wrapper || !els.length) {
      overlays[type].visible = false
      continue
    }

    // Offsets relative to the wrapper (scroll-independent: both rects are
    // viewport-relative, so the difference is the in-wrapper position).
    const base = wrapper.getBoundingClientRect().top
    const top = Math.min(...els.map((e) => e.getBoundingClientRect().top - base))
    const bottom = Math.max(
      ...els.map((e) => e.getBoundingClientRect().bottom - base)
    )
    overlays[type] = { top, height: bottom - top, visible: true }
  }
}

// Re-measure after the active set (or the item list) changes — flush:'post' so
// the DOM has settled — and pull the topmost active item into the rail's own
// scroll box (block:'nearest' never moves the page).
watch(
  [activeIds, headings],
  ([ids], [prevIds]) => {
    measure()
    const top = ids[0]
    if (top && top !== (prevIds as string[] | undefined)?.[0]) {
      liRefs.get(top)?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }
  },
  { flush: 'post' }
)

// The rail's width can change (wrap → item heights change), so the overlay must
// re-measure on resize even when the active set is unchanged.
const onResize = () => measure()
onMounted(() => window.addEventListener('resize', onResize, { passive: true }))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))
</script>

<template>
  <KunScrollShadow
    axis="vertical"
    shadow-size="3rem"
    class-name="scrollbar-hide max-h-[calc(100dvh-24rem)] overflow-y-scroll"
  >
    <template v-for="group in GROUPS" :key="group.type">
      <template v-if="headings.some((h) => h.type === group.type)">
        <h3 class="mb-2 text-sm font-medium">{{ group.title }}</h3>

        <div
          :ref="(el) => setWrapperRef(group.type, el)"
          class="relative mb-4"
        >
          <!-- Single highlight block for the whole visible run: a faint primary
               fill with all four corners rounded (no left bar). transition-all
               animates the grow/shrink as the run changes. -->
          <div
            v-show="overlays[group.type].visible"
            class="bg-primary/10 pointer-events-none absolute inset-x-0 rounded-md transition-all duration-300 ease-out"
            :style="{
              top: `${overlays[group.type].top}px`,
              height: `${overlays[group.type].height}px`
            }"
          />

          <ul class="relative z-10 space-y-1">
            <li
              v-for="item in headings.filter((h) => h.type === group.type)"
              :key="item.id"
              :ref="(el) => setLiRef(item.id, el)"
              :style="{ paddingLeft: `${(item.level - 1) * 0.75}rem` }"
            >
              <a
                :href="`#${item.id}`"
                @click.prevent="scrollToTOCElement(item.id)"
                :class="
                  cn(
                    'block py-1 pr-2 pl-2.5 text-sm transition-colors duration-300',
                    activeIds.includes(item.id)
                      ? 'text-primary font-medium'
                      : 'text-default-500 hover:text-primary'
                  )
                "
              >
                {{ item.text }}
              </a>
            </li>
          </ul>
        </div>
      </template>
    </template>
  </KunScrollShadow>
</template>
