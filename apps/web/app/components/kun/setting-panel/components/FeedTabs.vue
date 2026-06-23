<script setup lang="ts">
import { KUN_FEED_KIND_GROUPS, type KunFeedTab } from '~/constants/activity'

// Manage the home-feed tabs (设置 → 动态). Tabs live in the persisted settings
// store; the home feed renders them and sends each tab's kinds to the backend.
const settings = usePersistSettingsStore()
const { feedTabs } = storeToRefs(settings)

const addTab = () => {
  feedTabs.value.push({
    id: `tab-${Date.now().toString(36)}-${Math.floor(Math.random() * 1e6).toString(36)}`,
    name: '新标签',
    icon: 'lucide:tag',
    kinds: []
  })
}

const removeTab = (index: number) => {
  feedTabs.value.splice(index, 1)
}

// Reorder by splice-move so the persisted array (and the home tab order) updates.
const moveTab = (index: number, dir: -1 | 1) => {
  const target = index + dir
  if (target < 0 || target >= feedTabs.value.length) return
  const [item] = feedTabs.value.splice(index, 1)
  feedTabs.value.splice(target, 0, item!)
}

const toggleKind = (tab: KunFeedTab, kind: string) => {
  const i = tab.kinds.indexOf(kind)
  if (i >= 0) {
    tab.kinds.splice(i, 1)
  } else {
    tab.kinds.push(kind)
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-start justify-between gap-4">
      <div class="space-y-0.5">
        <p class="text-default-700 font-medium">动态标签</p>
        <p class="text-default-500 text-sm">
          自定义主页动态流的标签：增删、排序，并为每个标签勾选要展示的动态种类。
        </p>
      </div>
      <KunButton
        size="sm"
        variant="light"
        class="shrink-0"
        @click="settings.resetKUNGalgameFeedTabs()"
      >
        恢复默认
      </KunButton>
    </div>

    <div
      v-for="(tab, index) in feedTabs"
      :key="tab.id"
      class="border-default-200 space-y-3 rounded-lg border p-3"
    >
      <div class="flex items-center gap-2">
        <KunIcon
          :name="tab.icon || 'lucide:tag'"
          class="text-default-500 shrink-0 text-lg"
        />
        <KunInput v-model="tab.name" placeholder="标签名称" class="flex-1" />
        <KunButton
          size="sm"
          variant="light"
          :is-icon-only="true"
          :disabled="index === 0"
          @click="moveTab(index, -1)"
        >
          <KunIcon name="lucide:chevron-up" />
        </KunButton>
        <KunButton
          size="sm"
          variant="light"
          :is-icon-only="true"
          :disabled="index === feedTabs.length - 1"
          @click="moveTab(index, 1)"
        >
          <KunIcon name="lucide:chevron-down" />
        </KunButton>
        <KunButton
          size="sm"
          variant="light"
          color="danger"
          :is-icon-only="true"
          :disabled="feedTabs.length <= 1"
          @click="removeTab(index)"
        >
          <KunIcon name="lucide:trash-2" />
        </KunButton>
      </div>

      <KunInput
        v-model="tab.icon"
        placeholder="图标名称，如 lucide:gamepad-2"
        class="w-full"
      />

      <div class="space-y-2">
        <div
          v-for="group in KUN_FEED_KIND_GROUPS"
          :key="group.label"
          class="space-y-1.5"
        >
          <p class="text-default-400 text-xs">{{ group.label }}</p>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="kind in group.kinds"
              :key="kind.value"
              type="button"
              :class="
                cn(
                  'flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs transition-colors',
                  tab.kinds.includes(kind.value)
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-default-200 text-default-600 hover:bg-default-100'
                )
              "
              @click="toggleKind(tab, kind.value)"
            >
              <KunIcon :name="kind.icon" class="text-sm" />
              {{ kind.label }}
            </button>
          </div>
        </div>
      </div>

      <p v-if="!tab.kinds.length" class="text-warning text-xs">
        该标签未选择任何种类，将不会显示内容。
      </p>
    </div>

    <KunButton variant="light" class="w-full" @click="addTab">
      <KunIcon name="lucide:plus" />
      添加标签
    </KunButton>
  </div>
</template>
