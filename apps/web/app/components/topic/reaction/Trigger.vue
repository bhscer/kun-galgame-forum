<script setup lang="ts">
// The reaction trigger: an icon-only KunReaction that opens the picker popover.
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
    <!-- KunPopover wraps #trigger in an inline-block; a flex span here removes
         the inline line-box so the icon sits level with the sibling pills/chips
         (not nudged up by the baseline descender). -->
    <template #trigger>
      <span class="flex">
        <KunReaction :toggle="false" icon="lucide:smile-plus" label="表态" />
      </span>
    </template>

    <TopicReactionPicker :mine-keys="mineKeys" @select="onSelect" />
  </KunPopover>
</template>
