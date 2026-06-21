<script setup lang="ts">
// The reply editor (single milkdown body + submit row), extracted so
// TopicReplyPanel can host it in a bottom KunDrawer on mobile and in the fixed
// desktop panel — same logic, two shells. Multi-target tabs were retired: a
// reply is now one body. 「引用」-ed floors are shown as removable chips above the
// editor (their @mention / #quote tokens are folded into the body on publish),
// so the editor itself stays a plain empty composer.
const { isReplyRewriting, replyRewrite } = storeToRefs(useTempReplyStore())
const persistReplyStore = usePersistKUNGalgameReplyStore()
const { replyDraft, replyReferences } = storeToRefs(persistReplyStore)

const currentData = computed(() =>
  isReplyRewriting.value ? replyRewrite.value : replyDraft.value
)
</script>

<template>
  <div v-if="currentData" class="space-y-2">
    <!-- 「引用」 reference chips (create flow only; editing a reply keeps its
         existing inline tokens). -->
    <div
      v-if="!isReplyRewriting && replyReferences.length"
      class="flex flex-wrap items-center gap-2"
    >
      <span class="text-default-500 shrink-0 text-sm">回复</span>
      <span
        v-for="reference in replyReferences"
        :key="reference.replyId"
        class="bg-default-100 text-default-700 inline-flex items-center gap-1 rounded-full py-1 pr-1 pl-2.5 text-sm"
      >
        <span class="text-primary-500">@{{ reference.userName }}</span>
        <span class="text-default-400">#{{ reference.floor }}</span>
        <button
          type="button"
          class="text-default-400 hover:bg-default-200 hover:text-default-600 flex size-4 items-center justify-center rounded-full transition-colors"
          aria-label="移除引用"
          @click="persistReplyStore.removeReference(reference.replyId)"
        >
          <KunIcon name="lucide:x" class="size-3" />
        </button>
      </span>
    </div>

    <!-- Remount on context switch (draft ↔ editing a specific reply) so the
         editor loads the right body. -->
    <TopicReplyTargetEditor
      :key="isReplyRewriting ? `edit-${replyRewrite?.id}` : 'draft'"
      v-model="currentData.mainContent"
    />

    <TopicReplyPanelBtn class="mt-3" />
  </div>
</template>
