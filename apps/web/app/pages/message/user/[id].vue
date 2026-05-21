<script setup lang="ts">
definePageMeta({
  middleware: 'auth'
})

const route = useRoute()

const userId = parseInt((route.params as { id: string }).id)

useKunSeoMeta({
  title: '私信'
})
</script>

<template>
  <div class="h-full min-w-0 flex-1 pl-3">
    <ClientOnly>
      <MessagePmHeader :id="userId" />
    </ClientOnly>

    <MessagePmContainer :user-id="userId" />
  </div>
</template>
