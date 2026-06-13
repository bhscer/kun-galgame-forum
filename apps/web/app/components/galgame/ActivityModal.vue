<script setup lang="ts">
// Merged "编辑历史 + 更新请求" overlay launched from the galgame info
// card. Both views used to live as their own sidebar KunCards; folding
// them into one modal trades two persistent cards (low daily-use) for
// a single dismissible surface accessible via one button — keeps the
// info card lean while still surfacing the data when needed.
//
// The two children (`<GalgameHistory />` and `<GalgamePrContainer />`)
// are self-contained: they fetch their own data, manage their own
// pagination, and own their own UI. We just provide the KunTab
// switcher + KunModal frame.
type Tab = 'history' | 'pr'

const open = defineModel<boolean>({ required: true })
const props = withDefaults(
  defineProps<{
    // Tab to land on when the modal opens. The info-card button opens 编辑历史
    // (default); the pending-PR banner deep-links straight to 更新请求.
    initialTab?: Tab
  }>(),
  { initialTab: 'history' }
)

const activeTab = ref<Tab>(props.initialTab)

// Re-pin to initialTab on every open, so a banner-triggered open always lands
// on 更新请求 even if the user previously left the modal on 编辑历史.
watch(open, (isOpen) => {
  if (isOpen) {
    activeTab.value = props.initialTab
  }
})

const tabs = [
  { value: 'history' as const, textValue: '编辑历史', icon: 'lucide:history' },
  { value: 'pr' as const, textValue: '更新请求', icon: 'lucide:git-pull-request' }
]
</script>

<template>
  <KunModal v-model="open" inner-class-name="max-w-3xl w-[92vw]">
    <div class="space-y-4">
      <KunTab
        v-model="activeTab"
        :items="tabs"
        variant="underlined"
        color="primary"
        size="sm"
      />

      <div v-show="activeTab === 'history'">
        <GalgameHistory />
      </div>

      <div v-show="activeTab === 'pr'">
        <GalgamePrContainer />
      </div>
    </div>
  </KunModal>
</template>
