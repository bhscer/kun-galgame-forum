import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { EditStorePersist } from '~/store/types/edit/topic'

export const usePersistEditTopicStore = defineStore(
  'KUNGalgameEditTopic',
  () => {
    const mode = ref<EditStorePersist['mode']>('preview')
    const title = ref<EditStorePersist['title']>('')
    const content = ref<EditStorePersist['content']>('')
    const tags = ref<EditStorePersist['tags']>([])
    const category = ref<EditStorePersist['category']>('')
    const section = ref<EditStorePersist['section']>([])
    const isNSFW = ref<EditStorePersist['isNSFW']>(false)
    const coverImages = ref<EditStorePersist['coverImages']>([])

    const resetTopicData = () => {
      title.value = ''
      content.value = ''
      tags.value = []
      category.value = ''
      section.value = []
      isNSFW.value = false
      coverImages.value = []
    }

    return {
      mode,
      title,
      content,
      tags,
      category,
      section,
      isNSFW,
      coverImages,
      resetTopicData
    }
  },
  {
    persist: {
      storage: piniaPluginPersistedstate.localStorage()
    }
  }
)
