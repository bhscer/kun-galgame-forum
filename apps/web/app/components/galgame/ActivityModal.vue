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
const open = defineModel<boolean>({ required: true })

type Tab = 'history' | 'pr'
const activeTab = ref<Tab>('history')

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
