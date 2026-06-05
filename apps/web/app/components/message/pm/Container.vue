<script setup lang="ts">
import { replaceAsideItem } from '../aside/asideItemStore'

const props = defineProps<{
  userId: number
}>()

const historyContainer = ref<HTMLDivElement | null>(null)
const messageInput = ref('')
const messages = ref<ChatMessage[]>([])
const isLoadHistoryMessageComplete = ref(false)
const isSending = ref(false)
const isShowLoader = computed(() => {
  if (isLoadHistoryMessageComplete.value) {
    return false
  }
  if (messages.value.length < 30) {
    return false
  }
  return true
})
const currentUserId = usePersistUserStore().id
const userId = props.userId
const pageData = reactive({
  page: 1,
  limit: 30
})

const scrollToBottom = () => {
  if (historyContainer.value) {
    historyContainer.value.scrollTo({
      top: historyContainer.value.scrollHeight,
      behavior: 'smooth'
    })
  }
}

const getMessageHistory = async () => {
  const histories = await kunFetch<ChatMessage[]>('/message/chat/history', {
    method: 'GET',
    query: {
      receiverId: userId,
      page: pageData.page,
      limit: pageData.limit
    }
  })
  return Array.isArray(histories) ? histories : ([] as ChatMessage[])
}

const sendMessage = async () => {
  // Re-entry guard: the send can be triggered more than once near-simultaneously
  // (the input's @keydown.enter + the 发送 button, plus event bubbling). isSending
  // is flipped synchronously below before the await, so any concurrent/rapid
  // re-fire bails here instead of POSTing a duplicate — this is what caused a
  // single message to be inserted 2-3x.
  if (isSending.value) {
    return
  }
  if (!messageInput.value.trim()) {
    useMessage(10401, 'warn')
    return
  }
  if (messageInput.value.length > 1000) {
    useMessage(10402, 'warn')
    return
  }

  isSending.value = true
  const result = await kunFetch('/message/chat/send', {
    method: 'POST',
    body: { receiverId: userId, content: messageInput.value }
  })
  isSending.value = false

  if (result) {
    messageInput.value = ''
    // Reload latest messages to get the new one
    pageData.page = 1
    messages.value = await getMessageHistory()
    nextTick(() => scrollToBottom())
  }
}

// Sender-only recall: server validates ownership, but we still gate
// locally so we don't fire the request for foreign messages or already
// recalled ones (matches the canRecall guard inside Item.vue).
const handleRecallContextMenu = async (payload: {
  event: MouseEvent
  message: ChatMessage
}) => {
  const target = payload.message
  if (target.sender.id !== currentUserId || target.isRecall) {
    return
  }

  const confirmed = await useComponentMessageStore().alert(
    '撤回这条消息?',
    '撤回后对方将看到 “XX 撤回了一条消息”, 内容不可恢复'
  )
  if (!confirmed) {
    return
  }

  const ok = await kunFetch<string>('/message/chat/recall', {
    method: 'POST',
    body: { messageId: target.id }
  })
  if (!ok) {
    return
  }

  const idx = messages.value.findIndex((m) => m.id === target.id)
  if (idx !== -1) {
    messages.value[idx] = { ...messages.value[idx]!, isRecall: true }
  }
  useMessage('撤回成功', 'success')
}

const handleLoadHistoryMessages = async () => {
  if (!historyContainer.value) {
    return
  }

  const previousScrollHeight = historyContainer.value.scrollHeight
  const previousScrollTop = historyContainer.value.scrollTop

  pageData.page += 1
  const histories = await getMessageHistory()

  if (histories.length > 0) {
    messages.value.unshift(...histories)

    nextTick(() => {
      if (historyContainer.value) {
        const newScrollHeight = historyContainer.value.scrollHeight
        historyContainer.value.scrollTo({
          top: previousScrollTop + (newScrollHeight - previousScrollHeight)
        })
      }
    })
  } else {
    isLoadHistoryMessageComplete.value = true
  }
}

// Enter-to-send is bound on the input itself (@keydown.enter.prevent="sendMessage").
// A separate GLOBAL `window` keydown listener used to ALSO call sendMessage, so a
// single Enter fired the input handler AND the window handler (and any extra
// mounts duplicated it further) → 2-3 messages per send. It also sent on Enter
// when the chat input wasn't even focused. Removed; the scoped input handler is
// the single source of truth (plus the re-entry guard in sendMessage).
onMounted(async () => {
  messages.value = await getMessageHistory()

  nextTick(() => {
    scrollToBottom()
  })
})
</script>

<template>
  <div
    ref="historyContainer"
    class="h-[calc(100dvh-14rem)] overflow-y-auto py-3"
  >
    <div class="flex justify-center">
      <KunButton
        v-if="isShowLoader"
        @click="handleLoadHistoryMessages"
        size="sm"
        variant="light"
      >
        加载更多
      </KunButton>
    </div>

    <MessagePmItem
      v-for="message in messages"
      :key="message.id"
      :message="message"
      :is-sent="message.sender.id === currentUserId"
      @context-menu="handleRecallContextMenu"
    />

    <div v-if="!messages.length" class="text-default-500 py-10 text-center">
      暂无消息，发送一条消息开始聊天吧
    </div>
  </div>

  <div class="flex gap-2 border-t px-3 pt-3">
    <KunInput
      v-model="messageInput"
      placeholder="输入消息..."
      class="flex-1"
      @keydown.enter.prevent="sendMessage"
    />
    <KunButton @click="sendMessage" :loading="isSending"> 发送 </KunButton>
  </div>
</template>
