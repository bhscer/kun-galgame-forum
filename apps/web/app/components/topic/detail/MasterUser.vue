<script setup lang="ts">
const props = defineProps<{
  user: {
    id: number
    name: string
    avatar: string
    moemoepoint: number
  }
}>()

const user = computed(() => props.user)
</script>

<template>
  <!-- Desktop-only author rail. The sticky lives on THIS flex item (a direct child
       of Detail's row) so its containing block is the tall right column — that's
       what lets it stay put while the page scrolls. No card chrome; sticks at
       top-20 to clear the fixed top bar; height-capped so a long TOC scrolls. -->
  <div
    class="scrollbar-hide sticky top-20 hidden max-h-[calc(100dvh-6rem)] w-52 shrink-0 flex-col items-center gap-3 overflow-y-auto lg:flex"
  >
    <KunAvatar
      :disable-floating="true"
      class-name="w-46 h-46 hover:scale-100"
      size="original"
      image-class-name="w-46 h-46 shrink-0 rounded-lg"
      :user="user"
    />

    <KunLink
      underline="hover"
      :aria-label="props.user.name"
      :to="`/user/${user.id}/info`"
    >
      {{ user.name }}
    </KunLink>

    <p class="text-secondary flex items-center gap-1">
      <KunIcon class="text-inherit" name="lucide:lollipop" />
      {{ user.moemoepoint }}
    </p>

    <TopicDetailTableOfContent />
  </div>
</template>
