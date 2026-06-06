<script setup lang="ts">
const props = defineProps<{
  userId: number
}>()

const historyContainer = ref<HTMLDivElement | null>(null)
const messageInput = ref('')
const messages = ref<ChatMessage[]>([])
const isLoadHistoryMessageComplete = ref(false)
const isSending = ref(false)
const isUploadingImage = ref(false)
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
  if (isUploadingImage.value) {
    useMessage('图片正在上传中, 请稍候', 'warn')
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

// Paste or drop an image into the chat input → upload it to OUR image host
// (/image/message, the `message` preset) → append the markdown. Uploading is
// the only way an image survives the renderer's src allow-list (restricted to
// image.kungal.com — see RenderInline server-side); pasting an external image
// URL would just get stripped. Append-at-end is the right UX for attaching an
// image and leaves in-progress typing untouched.
const uploadAndAppendImages = async (files: File[]) => {
  const images = files.filter((file) => file.type.startsWith('image/'))
  if (!images.length) {
    return
  }

  isUploadingImage.value = true
  try {
    for (const image of images) {
      const formData = new FormData()
      formData.append('image', image)
      const url = await kunFetch<string>('/image/message', {
        method: 'POST',
        body: formData,
        watch: false
      })
      if (url) {
        const snippet = `![${image.name}](${url})`
        messageInput.value = messageInput.value
          ? `${messageInput.value} ${snippet}`
          : snippet
      }
    }
  } finally {
    isUploadingImage.value = false
  }
}

const handlePaste = (event: ClipboardEvent) => {
  const files = Array.from(event.clipboardData?.files ?? [])
  if (!files.some((file) => file.type.startsWith('image/'))) {
    return
  }
  // Only swallow the paste when it carries an image — plain-text pastes fall
  // through to the input untouched.
  event.preventDefault()
  uploadAndAppendImages(files)
}

const handleDrop = (event: DragEvent) => {
  const files = Array.from(event.dataTransfer?.files ?? [])
  if (!files.some((file) => file.type.startsWith('image/'))) {
    return
  }
  event.preventDefault()
  uploadAndAppendImages(files)
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
    class="h-[calc(100dvh-14rem)] space-y-3 overflow-y-auto py-3"
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

  <div class="border-t px-3 pt-3">
    <div v-if="isUploadingImage" class="text-default-500 mb-1 text-xs">
      图片上传中...
    </div>
    <div class="flex gap-2">
      <KunInput
        v-model="messageInput"
        placeholder="输入消息... (可粘贴或拖拽图片)"
        class="flex-1"
        @keydown.enter.prevent="sendMessage"
        @paste="handlePaste"
        @drop="handleDrop"
      />
      <KunButton @click="sendMessage" :loading="isSending"> 发送 </KunButton>
    </div>
  </div>
</template>
