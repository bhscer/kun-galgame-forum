<script setup lang="ts">
// Desktop sidebar icon rail: four groups (icon + label below). Hovering a group
// reveals a flyout menu of links to its right; focus-within mirrors hover for
// keyboard users. Mobile never renders this — it keeps the expanded drawer.
import { kunSidebarRail, type KunRailGroup } from '~/constants/layout'

const route = useRoute()

const openName = ref<string | null>(null)

const isLinkActive = (router: string) =>
  route.fullPath === router || route.fullPath.startsWith(`${router}/`)

// A group is active when the current route is its index or any of its links.
const isGroupActive = (group: KunRailGroup) => {
  if (group.router && isLinkActive(group.router)) return true
  return group.sections.some((s) => s.items.some((i) => isLinkActive(i.router)))
}
</script>

<template>
  <nav class="flex flex-col items-center gap-1">
    <div
      v-for="group in kunSidebarRail"
      :key="group.name"
      class="relative w-full"
      @mouseenter="openName = group.name"
      @mouseleave="openName = null"
      @focusin="openName = group.name"
      @focusout="openName = null"
    >
      <KunButton
        variant="light"
        color="default"
        :href="group.router"
        :aria-label="group.label"
        :class-name="
          cn(
            'h-auto w-full flex-col gap-1 py-2',
            isGroupActive(group) ? 'text-primary' : 'text-foreground'
          )
        "
      >
        <KunIcon :name="group.icon" class="text-xl" />
        <span class="text-[11px] leading-none">{{ group.label }}</span>
      </KunButton>

      <Transition name="rail-flyout">
        <div
          v-if="openName === group.name"
          class="absolute top-0 left-full z-50 pl-2"
        >
          <div
            class="kun-rail-flyout bg-content1 border-kun max-h-[80vh] min-w-52 overflow-y-auto rounded-lg border p-2 shadow-kun-sm"
          >
            <template
              v-for="(section, si) in group.sections"
              :key="`${group.name}-${si}`"
            >
              <div
                v-if="si > 0"
                class="border-default-200/60 my-1 border-t"
              />
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
  </nav>
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
