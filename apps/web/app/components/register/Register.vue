<script setup lang="ts">
// Per docs/integration/oauth/05-registration.md: registration is an
// identity-tier operation and must go through OAuth's hosted register
// page (oauth.kungal.com/auth/register). This component is the kungal-side
// jump button. It mirrors components/login/Login.vue exactly — same PKCE
// generation, same OAuth params, same sessionStorage keys — only the
// destination differs: /auth/register?redirect=<authorize-url> instead of
// /oauth/authorize. OAuth web registers + auto-logs-in the user, then
// chains into /oauth/authorize (where auto_consent=true skips the consent
// card for kungal), code lands on /auth/callback, kungal session created.
//
// The legacy in-app register form (own email-code flow → POST /api/user/register
// → write to local prisma DB) was deleted as part of the unified-registration
// migration.
const config = useRuntimeConfig()

const isLoading = ref(false)

const handleOAuthRegister = async () => {
  isLoading.value = true

  const codeVerifier = generateCodeVerifier()
  const codeChallenge = await generateCodeChallenge(codeVerifier)
  const state = generateState()

  sessionStorage.setItem('oauth_code_verifier', codeVerifier)
  sessionStorage.setItem('oauth_state', state)

  const authorizeParams = new URLSearchParams({
    client_id: config.public.oauthClientId as string,
    redirect_uri: config.public.oauthRedirectUri as string,
    response_type: 'code',
    scope: 'openid profile',
    state,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256'
  })

  // After registration completes, OAuth web window.location.href's to this
  // URL, which restarts the standard authorization-code flow on the now-
  // authenticated session. For first-party kungal (auto_consent=true) the
  // consent UI is silently skipped, code is issued, browser lands back on
  // kungal/auth/callback.
  const authorizeUrl = `${config.public.oauthServerUrl}/oauth/authorize?${authorizeParams}`

  const registerUrl = `${config.public.oauthFrontendUrl}/auth/register?redirect=${encodeURIComponent(authorizeUrl)}`
  window.location.href = registerUrl
}
</script>

<template>
  <div class="flex size-full items-center justify-center">
    <DocDetailBackgroundImage src="/login.webp" />

    <KunCard
      :is-transparent="false"
      :is-hoverable="false"
      class-name="w-88 p-8 select-none"
    >
      <div class="flex h-full flex-col justify-center">
        <div class="my-6">
          <h1 class="mb-6 flex items-center gap-2 text-2xl">
            <KunImage src="/favicon.webp" class-name="h-8 w-8 rounded-2xl" />
            注册
          </h1>
          <p class="text-default-500 mb-1">鲲的朋友! 很荣幸此生遇见你!</p>
          <p class="text-default-500">
            {{ `希望这个小家可以给你带来倾城美好!` }}
          </p>
        </div>

        <KunButton
          @click="handleOAuthRegister"
          :disabled="isLoading"
          class="bg-primary w-full rounded-3xl text-base tracking-wider text-white uppercase"
        >
          {{ isLoading ? '跳转中...' : '使用 鲲 Galgame OAuth 注册' }}
        </KunButton>

        <p class="text-default-400 mt-4 text-center text-xs">
          点击注册将跳转至鲲 Galgame OAuth 统一认证系统
        </p>

        <KunDivider class="my-4">
          <span class="mx-2">或</span>
        </KunDivider>

        <div class="flex flex-col gap-3 text-center">
          <KunLink to="/login">已有账号？登录</KunLink>
        </div>
      </div>
    </KunCard>
  </div>
</template>
