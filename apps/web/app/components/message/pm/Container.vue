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
// Images pasted/dropped into the input are uploaded immediately and held here
// as removable chips (NOT dumped as a raw `![](url)` into the text box). On
// send they're appended to the message as markdown — see sendMessage.
const pendingImages = ref<{ name: string; url: string }[]>([])
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
  // (the textarea's Enter handler + the 发送 button). isSending is flipped
  // synchronously below before the await, so any rapid re-fire bails here
  // instead of POSTing a duplicate.
  if (isSending.value) {
    return
  }
  if (isUploadingImage.value) {
    useMessage('图片正在上传中, 请稍候', 'warn')
    return
  }

  // Compose the wire content: trimmed text plus one markdown image per pending
  // chip (newline-joined so text and images stack in the rendered bubble).
  const text = messageInput.value.trim()
  const imageMarkdown = pendingImages.value
    // Strip markdown-breaking chars from the filename used as alt, so a name
    // like "a]b(c).png" can't terminate the ![alt](url) syntax early.
    .map((img) => `![${img.name.replace(/[[\]()]/g, '')}](${img.url})`)
    .join(' ')
  const content = [text, imageMarkdown].filter(Boolean).join('\n')

  if (!content) {
    useMessage(10401, 'warn')
    return
  }
  if (content.length > 1000) {
    useMessage(10402, 'warn')
    return
  }

  isSending.value = true
  const result = await kunFetch('/message/chat/send', {
    method: 'POST',
    body: { receiverId: userId, content }
  })
  isSending.value = false

  if (result) {
    messageInput.value = ''
    pendingImages.value = []
    // Reload latest messages to get the new one
    pageData.page = 1
    messages.value = await getMessageHistory()
    nextTick(() => scrollToBottom())
  }
}

// Upload pasted/dropped images to OUR image host (/image/message, the `message`
// preset) and hold each as a chip. Uploading is the only way an image survives
// the renderer's src allow-list (image.kungal.iloveren.link — see RenderInline
// server-side); a pasted external image URL would just get stripped.
const uploadImages = async (files: File[]) => {
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
        pendingImages.value.push({ name: image.name, url })
      }
    }
  } finally {
    isUploadingImage.value = false
  }
}

const removePendingImage = (index: number) => {
  pendingImages.value.splice(index, 1)
}

// paste/drop are bound ONCE on the input wrapper (not on the textarea), so a
// single paste triggers a single upload. Binding @paste on KunInput previously
// double-fired — KunInput re-applies $attrs to BOTH its root div and inner
// <input>, so the bubbling paste hit two listeners — which uploaded (and thus
// sent) the image twice.
const handlePaste = (event: ClipboardEvent) => {
  const files = Array.from(event.clipboardData?.files ?? [])
  if (!files.some((file) => file.type.startsWith('image/'))) {
    return
  }
  // Only swallow the paste when it carries an image — plain-text pastes fall
  // through to the textarea untouched.
  event.preventDefault()
  uploadImages(files)
}

const handleDrop = (event: DragEvent) => {
  uploadImages(Array.from(event.dataTransfer?.files ?? []))
}

// Enter sends; Shift+Enter inserts a newline. Guard isComposing so pressing
// Enter to confirm an IME (pinyin, etc.) candidate doesn't fire a send.
const handleEnter = (event: KeyboardEvent) => {
  if (event.isComposing || event.shiftKey) {
    return
  }
  event.preventDefault()
  sendMessage()
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

// Enter-to-send is handled on the textarea (@keydown.enter="handleEnter").
// A separate GLOBAL `window` keydown listener used to ALSO call sendMessage, so a
// single Enter fired the input handler AND the window handler (and any extra
// mounts duplicated it further) → 2-3 messages per send. It also sent on Enter
// when the chat input wasn't even focused. Removed; the scoped textarea handler is
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

  <div
    class="border-t px-3 pt-3"
    @paste="handlePaste"
    @drop.prevent="handleDrop"
    @dragover.prevent
  >
    <!--
      Pending image attachments as removable thumbnail chips, shown above the
      textarea instead of dumping a raw markdown URL into it. The dashed box is
      the in-flight upload indicator.
    -->
    <div
      v-if="pendingImages.length || isUploadingImage"
      class="mb-2 flex flex-wrap gap-2"
    >
      <div
        v-for="(img, index) in pendingImages"
        :key="img.url"
        class="border-default-200 relative h-16 w-16 overflow-hidden rounded-lg border"
      >
        <img :src="img.url" :alt="img.name" class="h-full w-full object-cover" />
        <button
          type="button"
          @click="removePendingImage(index)"
          class="bg-background/70 text-default-600 hover:text-danger absolute top-0.5 right-0.5 flex h-5 w-5 items-center justify-center rounded-full text-xs leading-none"
          aria-label="移除图片"
        >
          ✕
        </button>
      </div>
      <div
        v-if="isUploadingImage"
        class="border-default-200 text-default-500 flex h-16 w-16 items-center justify-center rounded-lg border border-dashed text-xs"
      >
        上传中...
      </div>
    </div>

    <div class="flex items-end gap-2">
      <KunTextarea
        v-model="messageInput"
        placeholder="输入消息... (可粘贴或拖拽图片, Enter 发送, Shift+Enter 换行)"
        class="flex-1"
        :auto-grow="true"
        :rows="1"
        max-height="160px"
        @keydown.enter="handleEnter"
      />
      <KunButton @click="sendMessage" :loading="isSending"> 发送 </KunButton>
    </div>
  </div>
</template>
