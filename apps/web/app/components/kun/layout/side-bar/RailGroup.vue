<script setup lang="ts">
// One desktop rail group: an icon tile whose hover reveals a flyout of links to
// its RIGHT. The hover logic is kun-ui 2.1's useKunPointerMenu — a coordinate
// safe-triangle (you can travel diagonally from the tile into the panel without
// it closing), open/close delays, and a shared `group` so the four tiles behave
// like a menu bar (instant sibling switch, only one open). Touch falls back to
// the link tap; no focus steal. Positioning stays custom because KunPopover only
// supports top/bottom, and the rail needs a right-side flyout.
import { useKunPointerMenu } from '@kungal/ui-vue'
import type { KunRailGroup } from '~/constants/layout'

const props = defineProps<{
  group: KunRailGroup
}>()

const route = useRoute()
const isLinkActive = (router: string) =>
  route.fullPath === router || route.fullPath.startsWith(`${router}/`)

const isGroupActive = computed(() => {
  if (props.group.router && isLinkActive(props.group.router)) return true
  return props.group.sections.some((s) =>
    s.items.some((i) => isLinkActive(i.router))
  )
})

const open = ref(false)
const panelRef = ref<HTMLElement | null>(null)
const { triggerHandlers, panelHandlers } = useKunPointerMenu(panelRef, {
  open,
  group: 'kun-sidebar-rail',
  closeDelay: 150
})
</script>

<template>
  <div class="relative w-full">
    <KunButton
      variant="light"
      color="default"
      :href="group.router"
      :aria-label="group.label"
      :class-name="
        cn(
          'h-auto w-full flex-col gap-1 py-2',
          isGroupActive ? 'text-primary' : 'text-foreground'
        )
      "
      @pointerenter="triggerHandlers.pointerenter"
      @pointerleave="triggerHandlers.pointerleave"
    >
      <KunIcon :name="group.icon" class="text-xl" />
      <span class="text-[11px] leading-none">{{ group.label }}</span>
    </KunButton>

    <Transition name="rail-flyout">
      <div
        v-if="open"
        ref="panelRef"
        class="absolute top-0 left-full z-50 pl-2"
        @pointerenter="panelHandlers.pointerenter"
        @pointerleave="panelHandlers.pointerleave"
      >
        <div
          class="bg-content1 border-kun max-h-[80vh] min-w-52 overflow-y-auto rounded-lg border p-2 shadow-kun-sm"
        >
          <template
            v-for="(section, si) in group.sections"
            :key="`${group.name}-${si}`"
          >
            <div v-if="si > 0" class="border-default-200/60 my-1 border-t" />
            <p
              v-if="section.label"
              class="text-default-500 px-2 pt-1 pb-0.5 text-xs select-none"
            >
              {{ section.label }}
            </p>
            <KunLink
              v-for="link in section.items"
              :key="link.router"
              :to="link.router"
              :target="link.external ? '_blank' : undefined"
              underline="none"
              color="default"
              :class-name="
                cn(
                  'hover:bg-default-100 flex items-center gap-2 rounded-md px-2 py-1.5 text-sm',
                  isLinkActive(link.router)
                    ? 'bg-accent text-primary'
                    : 'text-foreground'
                )
              "
            >
              <KunIcon
                v-if="link.icon"
                :name="link.icon"
                class="shrink-0 text-base"
              />
              <span class="whitespace-nowrap">{{ link.label }}</span>
              <span v-if="link.hint" class="text-primary ml-auto text-xs">
                {{ link.hint }}
              </span>
            </KunLink>
          </template>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.rail-flyout-enter-active,
.rail-flyout-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.rail-flyout-enter-from,
.rail-flyout-leave-to {
  opacity: 0;
  transform: translateX(-4px);
}
</style>
