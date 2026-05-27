<script setup lang="ts">
const route = useRoute()

const userId = computed(() => {
  return parseInt((route.params as { id: string }).id)
})

const { data } = await useKunFetch<UserInfo | 'banned'>(
  `/user/${userId.value}`,
  { query: { userId } }
)

if (data.value === 'banned') {
  // Banned profile: noindex (don't carry the now-removed user across
  // search-engine caches) but keep a minimal title so the page header
  // shows a sensible "已被封禁" hint while logged-in admins still see it.
  useKunDisableSeo('该用户已被封禁')
} else if (data.value) {
  useKunSeoMeta({
    title: data.value.name,
    description: data.value.bio
  })
} else {
  // Missing data → don't let `undefined` text leak into search results.
  useKunDisableSeo('未找到该用户')
}
</script>

<template>
  <div class="contents">
    <KunCard
      :is-hoverable="false"
      :is-transparent="false"
      class-name="m-auto"
      content-class="h-[calc(100dvh-120px)]"
      v-if="data !== 'banned'"
    >
      <div v-if="data" class="flex h-full w-full">
        <UserNavBar
          :user="{ id: data.id, name: data.name, avatar: data.avatar }"
        />

        <div class="scrollbar-hide h-full w-full overflow-y-auto pl-3">
          <NuxtPage :user="data" />
        </div>
      </div>

      <KunNull v-if="!data" description="未找到该用户" />
    </KunCard>

    <KunCard
      v-if="data === 'banned'"
      :is-hoverable="false"
      :is-transparent="false"
      content-class="h-[calc(100dvh-120px)]"
    >
      <KunNull description="此用户已被封禁" />
    </KunCard>
  </div>
</template>
