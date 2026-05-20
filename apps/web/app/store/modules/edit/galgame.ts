import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import { createEmptyLocaleMap, resetReactiveState } from '~/store/index'
import type { GalgameStorePersist } from '~/store/types/edit/galgame'

export const usePersistEditGalgameStore = defineStore(
  'KUNGalgameEditGalgame',
  () => {
    const vndbId = ref<GalgameStorePersist['vndbId']>('')
    const name = reactive<GalgameStorePersist['name']>({
      'en-us': '',
      'ja-jp': '',
      'zh-cn': '',
      'zh-tw': ''
    })
    const introduction = reactive<GalgameStorePersist['introduction']>({
      'en-us': '',
      'ja-jp': '',
      'zh-cn': '',
      'zh-tw': ''
    })
    const contentLimit = ref<GalgameStorePersist['contentLimit']>('sfw')
    // Wiki defaults: original_language=ja-jp, age_limit=r18. We default
    // age_limit to 'all' instead, because publishing R18 without the user
    // opting in is a content-policy risk on a default-SFW site (per audit
    // §10). User must consciously flip to r18 if applicable.
    const ageLimit = ref<GalgameStorePersist['ageLimit']>('all')
    const originalLanguage =
      ref<GalgameStorePersist['originalLanguage']>('ja-jp')
    const aliases = ref<GalgameStorePersist['aliases']>([])
    // U1: "" = unknown release date; serialized to wire `release_date` (a
    // bare empty string is valid per the schema's "empty OR YYYY-MM-DD"
    // refinement). TBA defaults false; user opts in for未发布 entries.
    const releaseDate = ref<GalgameStorePersist['releaseDate']>('')
    const releaseDateTBA = ref<GalgameStorePersist['releaseDateTBA']>(false)

    const resetEditGalgameStore = () => {
      vndbId.value = ''
      resetReactiveState(name, createEmptyLocaleMap())
      resetReactiveState(introduction, createEmptyLocaleMap())
      contentLimit.value = 'sfw'
      ageLimit.value = 'all'
      originalLanguage.value = 'ja-jp'
      aliases.value = []
      releaseDate.value = ''
      releaseDateTBA.value = false
    }

    return {
      vndbId,
      name,
      introduction,
      contentLimit,
      ageLimit,
      originalLanguage,
      aliases,
      releaseDate,
      releaseDateTBA,

      resetEditGalgameStore
    }
  },
  {
    persist: {
      storage: piniaPluginPersistedstate.localStorage()
    }
  }
)
