import { defineStore } from 'pinia'
import { ref } from 'vue'
import { withImageVariant } from '../../../shared/utils/getEffectiveBanner'
import type { UserStore } from '../types/user'

export const usePersistUserStore = defineStore(
  'KUNGalgameUser',
  () => {
    const id = ref<UserStore['id']>(0)
    const name = ref<UserStore['name']>('')
    const avatar = ref<UserStore['avatar']>('')
    const avatarMin = ref<UserStore['avatarMin']>('')
    const moemoepoint = ref<UserStore['moemoepoint']>(0)
    const role = ref<UserStore['role']>(0)
    const isCheckIn = ref<UserStore['isCheckIn']>(false)
    const dailyToolsetUploadBytes = ref<UserStore['dailyToolsetUploadBytes']>(0)

    const setUserInfo = (user: UserStore) => {
      id.value = user.id
      name.value = user.name
      avatar.value = user.avatar
      // withImageVariant picks the right separator per URL family:
      // image_service hash-addressed URLs get `_100`, legacy nitro
      // avatar paths get `-100`. Both coexist until the bulk migration.
      avatarMin.value = withImageVariant(user.avatar, '100')
      moemoepoint.value = user.moemoepoint
      role.value = user.role
      isCheckIn.value = user.isCheckIn
      dailyToolsetUploadBytes.value = user.dailyToolsetUploadBytes
    }

    const resetUser = () => {
      id.value = 0
      name.value = ''
      avatar.value = ''
      avatarMin.value = ''
      moemoepoint.value = 0
      role.value = 0
      isCheckIn.value = false
      dailyToolsetUploadBytes.value = 0
    }

    return {
      id,
      name,
      avatar,
      avatarMin,
      moemoepoint,
      role,
      isCheckIn,
      dailyToolsetUploadBytes,
      setUserInfo,
      resetUser
    }
  },
  {
    persist: true
  }
)
