<script setup lang="ts">
// The reaction trigger: an icon-only KunButton that opens the picker popover.
// State is injected (shared with the chips), so reacting here updates the bar.
const { toggle, mineKeys } = inject(reactionsKey)!

const picker = ref<{ close: () => void } | null>(null)

const onSelect = (key: string) => {
  picker.value?.close()
  toggle(key)
}
</script>

<template>
  <KunPopover ref="picker" position="top-start">
    <template #trigger>
      <KunReaction :toggle="false" icon="lucide:smile-plus" label="表态" />
    </template>

    <TopicReactionPicker :mine-keys="mineKeys" @select="onSelect" />
  </KunPopover>
</template>
