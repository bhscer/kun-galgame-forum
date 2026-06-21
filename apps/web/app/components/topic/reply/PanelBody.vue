<script setup lang="ts">
// The reply editor (single milkdown body + submit row), extracted so
// TopicReplyPanel can host it in a bottom KunDrawer on mobile and in the fixed
// desktop panel — same logic, two shells. Multi-target tabs were retired: a
// reply is now one body with inline @mention / #quote tokens (the 「引用」
// button appends those tokens to the draft).
const { isReplyRewriting, replyRewrite } = storeToRefs(useTempReplyStore())
const { replyDraft } = storeToRefs(usePersistKUNGalgameReplyStore())

const currentData = computed(() =>
  isReplyRewriting.value ? replyRewrite.value : replyDraft.value
)
</script>

<template>
  <div v-if="currentData" class="space-y-2">
    <!-- Remount on context switch (draft ↔ editing a specific reply) so the
         editor loads the right body; in-context appends from 「引用」 re-sync
         live via the editor's value watch. -->
    <TopicReplyTargetEditor
      :key="isReplyRewriting ? `edit-${replyRewrite?.id}` : 'draft'"
      v-model="currentData.mainContent"
    />

    <TopicReplyPanelBtn class="mt-3" />
  </div>
</template>
