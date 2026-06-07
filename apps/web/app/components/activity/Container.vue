<script setup lang="ts">
import {
  KUN_ACTIVITY_TYPE_TYPE,
  KUN_ACTIVITY_GROUPS,
  KUN_ACTIVITY_ICON_MAP
} from '~/constants/activity'

const pageData = reactive({
  page: 1,
  limit: 50,
  type: 'TOPIC_CREATION'
})

// Category picker is a grouped dropdown (18 flat tabs overflowed). Selecting a
// category resets pagination and closes the popover.
const categoryPopover = ref<{ close: () => void } | null>(null)
const selectCategory = (type: string) => {
  if (pageData.type !== type) {
    pageData.type = type
    pageData.page = 1
  }
  categoryPopover.value?.close()
}

// Global "显示没有下载资源的 Galgame" preference (cookie-persisted, SSR-safe).
// Off (default) drops resource-less galgames' creation rows; computed query
// keeps page/type reactive AND re-fetches when the toggle changes.
const settings = usePersistSettingsStore()
const { data, status } = await useKunFetch<{
  items: ActivityItem[]
  total: number
}>('/activity', {
  method: 'GET',
  query: computed(() => ({
    ...pageData,
    showNoResource: settings.showKUNGalgameNoResource
  }))
})
</script>

<template>
  <KunCard
    :is-transparent="false"
    v-if="data"
    content-class="space-y-3"
    :is-hoverable="false"
  >
    <KunHeader
      name="最新动态"
      description="这里展示了论坛的所有动态, 包括 Galgame, Galgame 资源, Galgame 网站, 话题, 回复, 评论, 网站更新 等"
    />

    <KunPopover
      ref="categoryPopover"
      position="bottom-start"
      inner-class="w-60 max-h-[70vh] overflow-y-auto p-1.5"
    >
      <template #trigger>
        <div
          class="border-default-200 hover:border-primary flex w-fit cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors"
        >
          <KunIcon
            :name="KUN_ACTIVITY_ICON_MAP[pageData.type]"
            class="text-primary"
          />
          <span class="font-medium">
            {{ KUN_ACTIVITY_TYPE_TYPE[pageData.type] }}
          </span>
          <KunIcon name="lucide:chevron-down" class="text-default-400 ml-1" />
        </div>
      </template>

      <div
        v-for="group in KUN_ACTIVITY_GROUPS"
        :key="group.label"
        class="mb-1 last:mb-0"
      >
        <div class="text-default-400 px-2 py-1 text-xs font-medium">
          {{ group.label }}
        </div>
        <button
          v-for="type in group.types"
          :key="type"
          type="button"
          @click="selectCategory(type)"
          :class="
            cn(
              'flex w-full cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 text-sm transition-colors',
              type === pageData.type
                ? 'bg-primary/10 text-primary'
                : 'text-foreground hover:bg-default-100'
            )
          "
        >
          <KunIcon :name="KUN_ACTIVITY_ICON_MAP[type]" class="shrink-0" />
          <span class="flex-1 text-left">
            {{ KUN_ACTIVITY_TYPE_TYPE[type] }}
          </span>
          <KunIcon
            v-if="type === pageData.type"
            name="lucide:check"
            class="text-primary shrink-0"
          />
        </button>
      </div>
    </KunPopover>

    <div
      v-for="(activity, index) in data.items"
      :key="index"
      class="flex flex-col space-y-2"
    >
      <KunLink
        underline="none"
        color="default"
        :to="activity.link"
        class-name="hover:text-primary line-clamp-3 break-all transition-colors"
      >
        {{ activity.content }}
        <KunChip class-name="cursor-pointer" color="primary" size="xs">
          {{ KUN_ACTIVITY_TYPE_TYPE[activity.type] }}
        </KunChip>
      </KunLink>

      <div class="flex items-center space-x-2">
        <KunUser size="sm" v-if="activity.actor" :user="activity.actor" />
        <span class="text-default-500 text-sm">
          <KunTime :time="activity.timestamp" />
        </span>
      </div>
    </div>

    <KunPagination
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.total / pageData.limit)"
      :is-loading="status === 'pending'"
    />
  </KunCard>
</template>
