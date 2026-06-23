<script setup lang="ts">
definePageMeta({
  middleware: 'auth'
})

const { messageStatus } = storeToRefs(useTempSettingStore())

useHead({ title: `消息 - ${kungal.titleShort}` })

onMounted(() => (messageStatus.value = 'online'))
</script>

<template>
  <div class="flex h-[calc(100dvh-120px)] flex-row">
    <MessageAsideContainer />
    <NuxtPage />
  </div>
</template>
