<script setup lang="ts">
definePageMeta({ layout: 'blank' })

const route = useRoute()
const error = ref('')

// OAuth callback is a transient redirect-only page — never something a
// search engine should index.
useKunDisableSeo('OAuth 登录回调')

onMounted(async () => {
  const code = route.query.code as string
  const returnedState = route.query.state as string
  const savedState = sessionStorage.getItem('oauth_state')
  const codeVerifier = sessionStorage.getItem('oauth_code_verifier')

  // Clean up
  sessionStorage.removeItem('oauth_state')
  sessionStorage.removeItem('oauth_code_verifier')

  if (!code) {
    error.value = '未收到授权码'
    redirectToLogin()
    return
  }

  if (returnedState !== savedState) {
    error.value = 'State 不匹配，可能存在安全风险'
    redirectToLogin()
    return
  }

  if (!codeVerifier) {
    error.value = 'PKCE 验证器丢失，请重新登录'
    redirectToLogin()
    return
  }

  // Matches BE dto.UserProfile exactly. Email is owned by OAuth and
  // NOT returned here — the frontend fetches it from OAuth's
  // /oauth/userinfo on demand (per the BE comment on UserProfile).
  const result = await kunFetch<{
    id: number
    name: string
    avatar: string
    role: number
    moemoepoint: number
    bio: string
  }>('/auth/oauth/callback', {
    method: 'POST',
    body: { code, code_verifier: codeVerifier }
  })

  if (result) {
    const userStore = usePersistUserStore()
    userStore.setUserInfo({
      id: result.id,
      name: result.name,
      avatar: result.avatar,
      // withImageVariant picks the right separator per URL family:
      // image_service hash-addressed URLs get `_100`, legacy nitro
      // paths still on image.kungal.com get `-100`. The legacy avatar
      // bulk migration is pending; until it lands both coexist for
      // active users.
      avatarMin: result.avatar ? withImageVariant(result.avatar, '100') : '',
      moemoepoint: result.moemoepoint,
      role: result.role,
      isCheckIn: false,
      dailyToolsetUploadCount: 0
    })

    useKunLoliInfo(`登录成功! 欢迎来到 ${kungal.name}`)
    await navigateTo('/')
  } else {
    error.value = '登录失败，请重试'
    redirectToLogin()
  }
})

// After a failed callback, drop the user back at the homepage. The
// top-bar 登录 button (KunAuthModal) is one click away — no point in
// keeping a /login page that itself just shows the same modal.
const redirectToLogin = () => {
  setTimeout(() => navigateTo('/'), 2000)
}
</script>

<template>
  <div class="flex size-full items-center justify-center">
    <div class="text-center">
      <p v-if="!error" class="text-lg">正在登录...</p>
      <p v-else class="text-danger">{{ error }}</p>
    </div>
  </div>
</template>
