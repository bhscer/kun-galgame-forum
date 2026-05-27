<script setup lang="ts">
const route = useRoute()

const userId = computed(() => {
  return parseInt((route.params as { id: string }).id)
})

const { data } = await useKunFetch<UserInfo>(`/user/${userId.value}`)

// Banned profiles get a stripped {id, name, status: 1} payload from
// the BE — there's no `'banned'` sentinel string, so the previous
// `data === 'banned'` branch was dead. status !== 0 is the canonical
// "not in good standing" signal.
const isBanned = computed(() => data.value && data.value.status !== 0)

if (isBanned.value) {
  useKunDisableSeo('该用户已被封禁')
} else if (data.value) {
  useKunSeoMeta({
    title: data.value.name,
    description: data.value.bio
  })
} else {
  useKunDisableSeo('未找到该用户')
}
</script>

<template>
  <div class="contents">
    <KunCard
      v-if="!isBanned"
      :is-hoverable="false"
      :is-transparent="false"
      class-name="m-auto"
      content-class="h-[calc(100dvh-120px)]"
    >
      <div v-if="data" class="flex h-full w-full">
        <UserNavBar
          :user="{ id: data.id, name: data.name, avatar: data.avatar }"
        />

        <div class="scrollbar-hide h-full w-full overflow-y-auto pl-3">
          <NuxtPage :user="data" />
        </div>
      </div>

      <KunNull v-else description="未找到该用户" />
    </KunCard>

    <KunCard
      v-else
      :is-hoverable="false"
      :is-transparent="false"
      content-class="h-[calc(100dvh-120px)]"
    >
      <KunNull description="此用户已被封禁" />
    </KunCard>
  </div>
</template>
