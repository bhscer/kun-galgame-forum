<script setup lang="ts">
import {
  FRIEND_LINK_CATEGORY_OPTIONS,
  FRIEND_LINK_STATUS_OPTIONS
} from '~/constants/friendLink'

// FriendLink / FriendLinkInput / FriendLinkCategory are auto-imported (shared/types).
const props = defineProps<{
  modelValue: boolean
  initialData?: FriendLink | null
  defaultCategory?: FriendLinkCategory
}>()

const emits = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: FriendLinkInput]
}>()

const isOpen = computed({
  get: () => props.modelValue,
  set: (v) => emits('update:modelValue', v)
})
const isEditing = computed(() => !!props.initialData?.id)

const getInitial = (): FriendLinkInput => {
  const d = props.initialData
  const base: FriendLinkInput = {
    category: d?.category ?? props.defaultCategory ?? 'galgame',
    name: d?.name ?? '',
    link: d?.link ?? '',
    description: d?.description ?? '',
    banner: d?.banner ?? '',
    status: d?.status ?? 'normal'
  }
  return d?.id ? { ...base, id: d.id } : base
}

const form = reactive<FriendLinkInput>(getInitial())
watch(
  () => props.modelValue,
  (open) => {
    if (open) Object.assign(form, getInitial())
  }
)

// Banner upload goes through the same image_service path as inline images
// (POST /image/topic → topic preset webp), returning a full CDN URL.
const isUploading = ref(false)
const handleBannerUpload = async (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  isUploading.value = true
  const fd = new FormData()
  fd.append('image', file)
  const url = await kunFetch<string>('/image/topic', {
    method: 'POST',
    body: fd,
    watch: false
  })
  isUploading.value = false
  if (url) {
    form.banner = url
    useMessage('图片上传成功', 'success')
  }
}

const handleSubmit = () => {
  if (!form.name.trim()) {
    useMessage('请填写友链名称', 'warn')
    return
  }
  if (!form.link.trim()) {
    useMessage('请填写友链地址', 'warn')
    return
  }
  emits('submit', { ...form })
  isOpen.value = false
}
</script>

<template>
  <KunModal
    :is-dismissable="false"
    v-model="isOpen"
    inner-class-name="max-w-2xl"
  >
    <form @submit.prevent>
      <h2 class="mb-4 text-xl font-bold">
        {{ isEditing ? '编辑友链' : '添加友链' }}
      </h2>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <KunInput v-model="form.name" label="名称" required />
        <KunInput
          v-model="form.link"
          label="链接 (URL)"
          required
          placeholder="https://..."
        />
        <KunSelect
          v-model="form.category"
          label="分类"
          :options="FRIEND_LINK_CATEGORY_OPTIONS"
        />
        <KunSelect
          v-model="form.status"
          label="状态"
          :options="FRIEND_LINK_STATUS_OPTIONS"
        />
        <KunTextarea
          v-model="form.description"
          label="描述"
          auto-grow
          show-char-count
          :maxlength="500"
          class-name="md:col-span-2"
        />

        <div class="md:col-span-2">
          <label class="mb-1 block text-sm font-medium">图标 / Banner</label>
          <div class="flex items-start gap-3">
            <KunImage
              v-if="form.banner"
              :src="form.banner"
              class="border-default-200 h-20 w-32 shrink-0 rounded-md border object-cover"
            />
            <div class="flex flex-1 flex-col gap-2">
              <input
                type="file"
                accept="image/*"
                :disabled="isUploading"
                @change="handleBannerUpload"
              />
              <KunInput v-model="form.banner" placeholder="或直接填图片 URL" />
              <span v-if="isUploading" class="text-default-500 text-sm">
                上传中...
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <KunButton variant="light" color="danger" @click="isOpen = false">
          取消
        </KunButton>
        <KunButton
          color="primary"
          :disabled="isUploading"
          @click="handleSubmit"
        >
          {{ isEditing ? '保存' : '添加' }}
        </KunButton>
      </div>
    </form>
  </KunModal>
</template>
