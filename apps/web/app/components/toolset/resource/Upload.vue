<script setup lang="ts">
import {
  MAX_SMALL_FILE_SIZE,
  MAX_LARGE_FILE_SIZE,
  USER_DAILY_UPLOAD_LIMIT,
  MOEMOEPOINT_SINGLE_MB_DIVISOR
} from '~/config/upload'
import {
  initToolsetUploadSchema,
  completeToolsetUploadSchema,
  abortToolsetUploadSchema
} from '~/validations/toolset'
import {
  KUN_GALGAME_TOOLSET_UPLOAD_STATUS_MAP,
  type KUN_GALGAME_TOOLSET_UPLOAD_STATUS_CONST
} from '~/constants/toolset'

const props = defineProps<{
  toolsetId: number
}>()

const emits = defineEmits<{
  onUploadSuccess: [ToolsetUploadResult]
  onClose: []
}>()

const MB = 1024 * 1024
// Matches the Go backend's upload_service.ChunkSize constant. Used for
// pre-flight progress math; the API does the authoritative slicing when
// it issues presigned URLs in its UploadLargeResponse.parts.
const LARGE_CHUNK_SIZE = 5 * MB
const UPLOAD_TRANSFER_FAILED = 'UPLOAD_TRANSFER_FAILED'
const DEFAULT_BINARY_CONTENT_TYPE = 'application/octet-stream'
type ToolsetUploadStatus =
  (typeof KUN_GALGAME_TOOLSET_UPLOAD_STATUS_CONST)[number]
type ToolsetUploadPart = {
  partNumber: number
  etag: string
}

// Browsers don't always populate File.type for archive formats (.7z, .rar
// in particular often come back empty). Fall back to a generic binary
// content-type so the API's required-field validator passes and so the
// presigned PUT later sets a sensible Content-Type header on S3.
const resolveContentType = (file: File): string =>
  file.type && file.type.length > 0 ? file.type : DEFAULT_BINARY_CONTENT_TYPE

const { moemoepoint, dailyToolsetUploadBytes, role } = storeToRefs(
  usePersistUserStore()
)
const fileInput = ref<HTMLInputElement>()
const selectedFile = ref<File | null>(null)

const progress = ref(0)
const isDragging = ref(false)
const uploadStatus = ref<ToolsetUploadStatus>('idle')

const isLarge = computed(() => {
  const f = selectedFile.value
  return !!f && f.size > MAX_SMALL_FILE_SIZE
})
const isAdmin = computed(() => role.value > 1)
const dailyUploadLimit = computed(() => {
  if (isAdmin.value) {
    return MAX_LARGE_FILE_SIZE
  }

  // Remaining daily budget = (100MB + moemoepoint·MB) − bytes already used today.
  return Math.max(
    0,
    USER_DAILY_UPLOAD_LIMIT +
      moemoepoint.value * MB -
      dailyToolsetUploadBytes.value
  )
})
const maxSingleFileLimit = computed(() => {
  if (isAdmin.value) {
    return MAX_LARGE_FILE_SIZE
  }

  const moemoepointMaxSingleFile =
    Math.floor(moemoepoint.value / MOEMOEPOINT_SINGLE_MB_DIVISOR) * MB

  return Math.min(
    Math.max(USER_DAILY_UPLOAD_LIMIT, moemoepointMaxSingleFile),
    MAX_LARGE_FILE_SIZE
  )
})

const statusMessage = computed(() => {
  if (uploadStatus.value === 'largeUploading') {
    return `正在上传大文件【进度 ${progress.value}%】`
  } else {
    return KUN_GALGAME_TOOLSET_UPLOAD_STATUS_MAP[uploadStatus.value]
  }
})

const resetUploadState = () => {
  progress.value = 0
  uploadStatus.value = 'idle'
}

const setSelectedUploadFile = (file: File) => {
  selectedFile.value = file
  resetUploadState()
}

const throwIfUploadFailed = (response: Response) => {
  if (!response.ok) {
    throw new Error(UPLOAD_TRANSFER_FAILED)
  }
}

const isUploadTransferFailedError = (error: unknown) => {
  return error instanceof Error && error.message === UPLOAD_TRANSFER_FAILED
}

const notifyUploadTransferError = (error: unknown) => {
  if (isUploadTransferFailedError(error)) {
    useMessage('文件传输失败，请重试', 'error')
  }
}

const abortLargeUpload = async (upload: ToolsetLargeFileUploadResponse) => {
  const abortUploadData = { salt: upload.salt }
  const isValidAbortUploadData = useKunSchemaValidator(
    abortToolsetUploadSchema,
    abortUploadData
  )
  if (!isValidAbortUploadData) {
    return
  }

  try {
    await kunFetch(`/toolset/${props.toolsetId}/upload/abort`, {
      method: 'POST',
      body: abortUploadData
    })
  } catch (abortError) {
    console.error('Failed to abort toolset upload:', abortError)
  }
}

const checkFileValid = (file: File | null) => {
  if (!file) {
    return false
  }
  if (!isValidArchive(file.name || '')) {
    useMessage('我们仅支持 .7z, .zip, .rar 压缩格式上传', 'warn')
    return false
  }
  if (file.size > MAX_LARGE_FILE_SIZE) {
    useMessage(
      `文件大小超过最大文件限制 ${MAX_LARGE_FILE_SIZE / MB} MB`,
      'warn'
    )
    return false
  }
  if (file.size > dailyUploadLimit.value) {
    useMessage(
      `超出当日可用上传额度, 剩余 ${(dailyUploadLimit.value / MB).toFixed(2)} MB`,
      'warn'
    )
    return false
  }
  if (file.size > maxSingleFileLimit.value) {
    useMessage(
      `单文件大小超过限制, 最大 ${(maxSingleFileLimit.value / MB).toFixed(2)} MB`,
      'warn'
    )
    return false
  }
  return true
}

const pick = () => fileInput.value?.click()
const onChange = (e: Event) => {
  const t = e.target as HTMLInputElement
  const targetFile = t.files && t.files[0] ? t.files[0] : null
  const res = checkFileValid(targetFile)
  if (!res) {
    return
  }
  if (!targetFile) {
    return
  }
  setSelectedUploadFile(targetFile)
}
const onDrop = (e: DragEvent) => {
  e.preventDefault()
  e.stopPropagation()
  isDragging.value = false
  const dt = e.dataTransfer
  if (dt?.files && dt.files.length > 0) {
    const res = checkFileValid(dt.files[0]!)
    if (!res) {
      return
    }
    setSelectedUploadFile(dt.files[0]!)
  }
}
const onDragOver = (e: DragEvent) => {
  e.preventDefault()
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = 'copy'
  }
}
const onDragEnter = () => {
  isDragging.value = true
}
const onDragLeave = () => {
  isDragging.value = false
}
const clearSelected = () => {
  selectedFile.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
  resetUploadState()
}

const uploadSmall = async (f: File) => {
  uploadStatus.value = 'smallInit'
  const contentType = resolveContentType(f)
  const initUploadData = {
    toolsetId: props.toolsetId,
    filename: f.name,
    filesize: f.size,
    contentType
  }
  const isValidInitUploadData = useKunSchemaValidator(
    initToolsetUploadSchema,
    initUploadData
  )
  if (!isValidInitUploadData) {
    return
  }

  const initRes = await kunFetch<ToolsetSmallFileUploadResponse>(
    `/toolset/${props.toolsetId}/upload/small`,
    {
      method: 'POST',
      body: initUploadData
    }
  )
  if (!initRes) {
    uploadStatus.value = 'idle'
    return
  }

  let isUploadComplete = false
  try {
    uploadStatus.value = 'smallUploading'
    // Content-Type here must match the value sent during init — S3's
    // presigned URL is signed against it.
    const uploadRes = await fetch(initRes.presignedUrl, {
      headers: { 'Content-Type': contentType },
      method: 'PUT',
      body: f
    })
    throwIfUploadFailed(uploadRes)

    uploadStatus.value = 'smallComplete'
    const completeUploadData = { salt: initRes.salt }
    const isValidCompleteUploadData = useKunSchemaValidator(
      completeToolsetUploadSchema,
      completeUploadData
    )
    if (!isValidCompleteUploadData) {
      return
    }

    const done = await kunFetch<ToolsetUploadCompleteResponse>(
      `/toolset/${props.toolsetId}/upload/complete`,
      {
        method: 'POST',
        body: completeUploadData
      }
    )
    if (done) {
      useMessage('上传成功', 'success')
      emits('onUploadSuccess', {
        salt: initRes.salt,
        key: done.key,
        size: done.size
      })
      progress.value = 100
      uploadStatus.value = 'complete'
      isUploadComplete = true
    }
  } catch (error) {
    notifyUploadTransferError(error)
  } finally {
    if (!isUploadComplete) {
      resetUploadState()
    }
  }
}

const uploadLarge = async (f: File) => {
  uploadStatus.value = 'largeInit'
  const contentType = resolveContentType(f)
  const initUploadData = {
    toolsetId: props.toolsetId,
    filename: f.name,
    filesize: f.size,
    contentType
  }
  const isValidInitUploadData = useKunSchemaValidator(
    initToolsetUploadSchema,
    initUploadData
  )
  if (!isValidInitUploadData) {
    return
  }

  progress.value = 0
  const initRes = await kunFetch<ToolsetLargeFileUploadResponse>(
    `/toolset/${props.toolsetId}/upload/large`,
    {
      method: 'POST',
      body: initUploadData
    }
  )
  if (!initRes) {
    uploadStatus.value = 'idle'
    return
  }

  let isUploadComplete = false
  try {
    const partUrls = initRes.parts
    const parts: ToolsetUploadPart[] = []

    uploadStatus.value = 'largeUploading'
    for (let i = 0; i < partUrls.length; i++) {
      const currentPart = partUrls[i]
      if (!currentPart) {
        throw new Error('Missing upload part')
      }

      const { partNumber, presignedUrl } = currentPart
      const start = (partNumber - 1) * LARGE_CHUNK_SIZE
      const end = Math.min(start + LARGE_CHUNK_SIZE, f.size)
      const blob = f.slice(start, end)
      const resp = await fetch(presignedUrl, {
        headers: { 'Content-Type': contentType },
        method: 'PUT',
        body: blob
      })
      throwIfUploadFailed(resp)
      const etag = resp.headers.get('ETag') || resp.headers.get('etag')
      if (!etag) {
        throw new Error('Missing ETag')
      }
      parts.push({ partNumber, etag })
      progress.value = Math.round(((i + 1) / partUrls.length) * 100)
    }

    uploadStatus.value = 'largeComplete'
    const completeUploadData = {
      salt: initRes.salt,
      parts
    }
    const isValidCompleteUploadData = useKunSchemaValidator(
      completeToolsetUploadSchema,
      completeUploadData
    )
    if (!isValidCompleteUploadData) {
      return
    }

    const done = await kunFetch<ToolsetUploadCompleteResponse>(
      `/toolset/${props.toolsetId}/upload/complete`,
      {
        method: 'POST',
        body: completeUploadData
      }
    )
    if (done) {
      useMessage('上传成功', 'success')
      emits('onUploadSuccess', {
        salt: initRes.salt,
        key: done.key,
        size: done.size
      })
      uploadStatus.value = 'complete'
      isUploadComplete = true
    }
  } catch (error) {
    if (initRes?.uploadId) {
      await abortLargeUpload(initRes)
    }

    notifyUploadTransferError(error)
  } finally {
    if (!isUploadComplete) {
      resetUploadState()
    }
  }
}

const submit = async () => {
  const f = selectedFile.value
  if (!f) {
    useMessage('请选择文件', 'warn')
    return
  }
  if (f.size > MAX_SMALL_FILE_SIZE) {
    await uploadLarge(f)
  } else {
    await uploadSmall(f)
  }
}
</script>

<template>
  <div class="space-y-4">
    <input ref="fileInput" type="file" hidden @change="onChange" />

    <KunCard :is-hoverable="false" :is-transparent="true">
      <div
        class="cursor-pointer rounded-lg border-2 border-dashed p-6 text-center transition-colors"
        :class="
          cn(
            isDragging
              ? 'border-primary-500 bg-primary-50/50'
              : 'border-default-300 hover:border-default-500'
          )
        "
        @click="pick"
        @drop="onDrop"
        @dragover="onDragOver"
        @dragenter="onDragEnter"
        @dragleave="onDragLeave"
      >
        <div v-if="!selectedFile" class="flex flex-col items-center gap-2">
          <KunIcon
            name="lucide:upload-cloud"
            class="text-default-500 text-3xl"
          />
          <div class="text-default-600">点击或拖拽文件到此处</div>
        </div>

        <div v-else class="flex flex-col justify-center gap-2">
          <div class="flex items-center gap-3">
            <KunIcon
              name="lucide:file-check"
              class="text-success-600 text-xl"
            />
            <div class="text-default-700 font-medium">
              {{ selectedFile?.name }}
            </div>
          </div>

          <div class="flex items-center gap-3">
            <div class="text-default-500 text-xs">
              {{ formatFileSize(selectedFile!.size) }}
            </div>
            <span
              class="border-default-200 bg-default-100 text-default-600 rounded-full border px-2 py-0.5 text-xs"
            >
              {{
                isLarge
                  ? `文件大于 ${MAX_SMALL_FILE_SIZE / MB}MB, 分片上传`
                  : `文件小于 ${MAX_SMALL_FILE_SIZE / MB}MB, 直接上传`
              }}
            </span>
          </div>

          <KunProgress :value="progress" />

          <div
            class="text-default-500 flex items-center justify-center gap-2 text-sm"
          >
            <span>{{ statusMessage }}</span>
            <KunIcon
              class="text-sm"
              v-if="uploadStatus !== 'idle' && uploadStatus !== 'complete'"
              name="svg-spinners:90-ring-with-bg"
            />
            <KunIcon
              class="text-success-600 text-sm"
              v-if="uploadStatus === 'complete'"
              name="lucide:circle-check-big"
            />
          </div>
        </div>
      </div>
    </KunCard>

    <div class="flex items-center justify-end gap-2">
      <KunButton
        v-if="selectedFile"
        variant="light"
        color="danger"
        @click.stop="clearSelected"
      >
        移除文件
      </KunButton>
      <KunButton
        :loading="uploadStatus !== 'idle' && uploadStatus !== 'complete'"
        :disabled="!selectedFile || uploadStatus === 'complete'"
        @click="submit"
      >
        确认上传
      </KunButton>
    </div>
  </div>
</template>
