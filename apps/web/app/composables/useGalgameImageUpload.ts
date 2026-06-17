import type { Ref } from 'vue'
import type {
  GalgameImageUploadPreset,
  UploadGalgameImageResult
} from '~/utils/uploadGalgameImage'

interface ImageRowBase {
  image_hash: string
  sort_order: number
}

// Shared cover/screenshot batch uploader (K-PR3b image-first editor).
//
// Handles multi-file + drag-drop: each picked/dropped file is uploaded to
// image_service via /image/galgame, de-duplicated by content hash (a galgame
// can't list the same image twice; image_service dedups by content, so the
// SAME file re-uploaded returns the SAME hash), then appended as a new row at
// the next sort_order. Uploads run SEQUENTIALLY so `progressText` (n/total) is
// meaningful and a burst of files can't hammer the upload endpoint at once.
//
// The row shape differs (covers vs screenshots, which add `caption`), so the
// caller supplies `makeRow(res, sortOrder)`; everything else is shared.
export const useGalgameImageUpload = <T extends ImageRowBase>(opts: {
  preset: GalgameImageUploadPreset
  rows: Ref<T[]>
  makeRow: (res: UploadGalgameImageResult, sortOrder: number) => T
  dedupeLabel: string // '封面' | '截图', for the "skipped N duplicates" toast
}) => {
  const isUploading = ref(false)
  const done = ref(0)
  const total = ref(0)

  const progressText = computed(() =>
    total.value > 1 ? `上传中 ${done.value}/${total.value}` : '上传中'
  )

  const nextSortOrder = () =>
    opts.rows.value.reduce((m, r) => (r.sort_order > m ? r.sort_order : m), -1) +
    1

  const uploadFiles = async (files: File[]) => {
    const images = files.filter((f) => f.type.startsWith('image/'))
    if (!images.length) return

    isUploading.value = true
    done.value = 0
    total.value = images.length
    let duplicates = 0

    for (const file of images) {
      const res = await uploadGalgameImage(file, opts.preset, file.name)
      done.value++
      if (!res) continue // kunFetch already toasted the wiki error
      if (opts.rows.value.some((r) => r.image_hash === res.hash)) {
        duplicates++
        continue
      }
      opts.rows.value = [...opts.rows.value, opts.makeRow(res, nextSortOrder())]
    }

    isUploading.value = false
    total.value = 0
    if (duplicates > 0) {
      useMessage(`已跳过 ${duplicates} 张重复${opts.dedupeLabel}`, 'warn')
    }
  }

  return { isUploading, progressText, uploadFiles }
}
