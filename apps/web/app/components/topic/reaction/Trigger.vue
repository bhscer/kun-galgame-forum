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
      <KunButton
        :is-icon-only="true"
        variant="light"
        color="default"
        size="lg"
      >
        <KunIcon name="lucide:smile-plus" />
      </KunButton>
    </template>

    <TopicReactionPicker :mine-keys="mineKeys" @select="onSelect" />
  </KunPopover>
</template>
