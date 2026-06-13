<script setup lang="ts">
const { showKUNGalgameBackLoli } = storeToRefs(usePersistSettingsStore())
const { showKUNGalgamePanel } = storeToRefs(useTempSettingStore())

// The panel grew too many controls for one column, so everything is grouped
// into categories. Desktop shows the categories as a LEFT rail; mobile stacks
// them on TOP. KunTab's orientation is a prop (not responsive), so we render
// two instances bound to the same `activeTab`. The 差分图 (Loli) stays on the
// right on desktop and hides on mobile (no room) — gated in its own component.
const settingTabs = [
  { value: 'appearance', textValue: '外观', icon: 'lucide:palette' },
  { value: 'background', textValue: '背景', icon: 'lucide:image' },
  { value: 'content', textValue: '内容', icon: 'lucide:shield-alert' },
  { value: 'general', textValue: '通用', icon: 'lucide:settings-2' }
]
const activeTab = ref('appearance')
</script>

<template>
  <KunModal
    :model-value="showKUNGalgamePanel"
    @update:model-value="(value) => (showKUNGalgamePanel = value)"
    inner-class-name="overflow-visible w-[92vw] sm:max-w-2xl"
  >
    <div class="space-y-4">
      <div class="flex items-center gap-2 text-lg">
        <span>设置面板</span>

        <KunTooltip class-name="flex" text="设置面板帮助" position="bottom">
          <KunLink
            to="/doc/setting-panel-help"
            color="default"
            class="hover:text-primary"
          >
            <KunIcon name="lucide:circle-help" />
          </KunLink>
        </KunTooltip>
      </div>

      <!-- Mobile: category tabs on top. Wrapped in a plain div for the
           responsive show/hide — KunTab sets its own display, which would
           otherwise win over a `hidden`/`sm:hidden` utility put directly on it. -->
      <div class="sm:hidden">
        <KunTab
          v-model="activeTab"
          :items="settingTabs"
          variant="light"
          color="primary"
          size="sm"
          scrollable
        />
      </div>

      <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
        <!-- Desktop: category tabs as a left rail (same wrapper rationale). -->
        <div class="hidden shrink-0 sm:block sm:w-24">
          <KunTab
            v-model="activeTab"
            :items="settingTabs"
            orientation="vertical"
            variant="light"
            color="primary"
            full-width
          />
        </div>

        <div class="min-w-0 flex-1 space-y-4 sm:min-h-75">
          <div v-show="activeTab === 'appearance'" class="space-y-4">
            <KunSettingPanelComponentsMode />
            <KunSettingPanelComponentsConfigItems />
          </div>

          <div v-show="activeTab === 'background'">
            <KunSettingPanelComponentsBackground />
          </div>

          <div v-show="activeTab === 'content'">
            <KunSettingPanelComponentsNSFW />
          </div>

          <div v-show="activeTab === 'general'" class="space-y-4">
            <div class="flex items-center justify-between">
              显示琥珀
              <KunTooltip
                class-name="flex"
                text="是否显示网站右下角这只可爱的孩子"
                position="bottom"
              >
                <KunSwitch v-model="showKUNGalgameBackLoli" />
              </KunTooltip>
            </div>

            <KunSettingPanelComponentsReset />
          </div>
        </div>

        <!-- 差分图 — desktop only (the component is hidden sm:block) -->
        <KunSettingPanelComponentsLoli />
      </div>
    </div>
  </KunModal>
</template>
