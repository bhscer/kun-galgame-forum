<script setup lang="ts">
const props = defineProps<{
  topicId?: number
  targetUserId: number
  upvoteCount: number
  isUpvoted: boolean
  // Render as a left-justified labeled row for the ⋯ overflow menu.
  menu?: boolean
}>()

const { id, moemoepoint } = usePersistUserStore()
const isUpvoted = ref(props.isUpvoted)
const upvoteCount = ref(props.upvoteCount)
const pending = ref(false)

// The 推话题 dialog: an optional 一句话 (<=30 chars) "why I pushed it". Promise-
// based so the KunReaction / menu paths can await the user's choice. Backdrop /
// Esc / 取消 resolve as cancelled (null); 确定 resolves the trimmed text.
const dialogOpen = ref(false)
const description = ref('')
let resolver: ((value: string | null) => void) | null = null

// Cap at 30 characters as the user types (rune-safe; the backend caps too).
watch(description, (v) => {
  const runes = [...v]
  if (runes.length > 30) description.value = runes.slice(0, 30).join('')
})

const openDialog = (): Promise<string | null> => {
  description.value = ''
  dialogOpen.value = true
  return new Promise((resolve) => (resolver = resolve))
}
const confirmDialog = () => {
  const r = resolver
  resolver = null
  dialogOpen.value = false
  r?.(description.value.trim())
}
// Any other close (backdrop / Esc / 取消) resolves as cancelled.
watch(dialogOpen, (open) => {
  if (!open && resolver) {
    const r = resolver
    resolver = null
    r(null)
  }
})

// Shared: self / moemoepoint checks → the dialog → API. Returns true once the
// upvote lands. One-way (no undo); costs 10 萌萌点 and awards the author 5.
const confirmAndUpvote = async (): Promise<boolean> => {
  if (id === props.targetUserId) {
    useMessage(10241, 'warn')
    return false
  }
  if (moemoepoint < 10) {
    useMessage(10242, 'warn')
    return false
  }
  const desc = await openDialog()
  if (desc === null) return false

  pending.value = true
  const result = await kunFetch<string>(`/topic/${props.topicId}/upvote`, {
    method: 'PUT',
    body: { description: desc }
  })
  pending.value = false
  if (!result) return false

  useMessage(10238, 'success')
  return true
}

// Both the menu button + the reaction icon funnel here. Repeatable — a topic can
// be pushed again and again — so there's NO "already upvoted" guard; every click
// is a fresh push (each costs 10 萌萌点 + credits the author 5).
const handleClickUpvote = async () => {
  if (!id) {
    useAuthModal().open()
    return
  }
  if (pending.value) return
  if (await confirmAndUpvote()) {
    upvoteCount.value++
    isUpvoted.value = true
  }
}
</script>

<template>
  <KunButton
    v-if="menu"
    :variant="isUpvoted ? 'flat' : 'light'"
    :color="isUpvoted ? 'secondary' : 'default'"
    size="sm"
    class-name="w-full justify-start gap-2 whitespace-nowrap"
    :disabled="pending"
    @click="handleClickUpvote"
  >
    <KunIcon class-name="text-lg" name="lucide:sparkles" />
    推话题
    <span v-if="upvoteCount" class="text-default-500 ml-auto">
      {{ upvoteCount }}
    </span>
  </KunButton>

  <KunTooltip v-else text="推话题">
    <!-- Action mode (repeatable): each click is a fresh push, never disabled
         after upvoting. The count rolls when a push lands. -->
    <KunReaction
      :toggle="false"
      :count="upvoteCount"
      :disabled="pending"
      label="推话题"
      @click="handleClickUpvote"
    >
      <template #icon>
        <KunIcon name="lucide:sparkles" class="text-warning" />
      </template>
    </KunReaction>
  </KunTooltip>

  <KunModal v-model="dialogOpen" role="alertdialog" inner-class-name="max-w-md">
    <div class="space-y-4">
      <h3 class="text-lg font-medium">确定推这个话题吗？</h3>
      <p class="text-default-500 text-sm">
        推话题将消耗您
        <span class="text-warning-600 font-medium">10</span> 萌萌点，并给被推者增加
        <span class="text-success font-medium">5</span> 萌萌点。
      </p>
      <div>
        <KunInput
          v-model="description"
          placeholder="（可选）一句话，说说为什么推它 ✨"
        />
        <p class="text-default-400 mt-1 text-right text-xs">
          {{ [...description].length }}/30
        </p>
      </div>
      <div class="flex justify-end gap-3">
        <KunButton variant="light" color="default" @click="dialogOpen = false">
          取消
        </KunButton>
        <KunButton color="secondary" @click="confirmDialog">确定推</KunButton>
      </div>
    </div>
  </KunModal>
</template>
