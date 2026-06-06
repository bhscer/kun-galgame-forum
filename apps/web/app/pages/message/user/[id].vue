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
  <!--
    flex column bounded by the parent's h-[calc(100dvh-120px)] (see message.vue).
    The container's history flexes to fill and SHRINKS as the input grows, so a
    tall composer (chips + multi-line text) never pushes past the viewport.
  -->
  <div class="flex h-full min-w-0 flex-1 flex-col pl-3">
    <ClientOnly>
      <MessagePmHeader :id="userId" />
    </ClientOnly>

    <MessagePmContainer :user-id="userId" />
  </div>
</template>
