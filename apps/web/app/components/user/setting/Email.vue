<script setup lang="ts">
// Email change moved out of kungal entirely (docs/oauth/02-user-profile
// .md: "修改 email 不在这里 —— email 必须走 /auth/email/send-code
// + /auth/email"). The old in-app two-step flow is gone; users jump
// to the OAuth account-center profile page, which owns the
// send-code + confirm pipeline. Once they finish there, the
// usual session-refresh cycle picks up the new email at next sign-in.
//
// Resolved exactly like Avatar.vue does for the legacy avatar jump:
// strip /api/vN/ off the configured OAuth server URL to land on the
// account-center root, then append /profile.
const oauthProfileURL = computed(() => {
  const apiBase = useRuntimeConfig().public.oauthServerUrl || ''
  return apiBase.replace(/\/api\/v\d+\/?$/, '') + '/profile'
})
</script>

<template>
  <KunCard :is-hoverable="false" content-class="space-y-3">
    <div class="space-y-2">
      <span class="text-xl">更改邮箱</span>
      <p class="text-default-500 text-sm">
        邮箱由 OAuth 账户中心统一管理。修改邮箱需要二次验证 (邮箱验证码),
        请前往 OAuth 账户中心进行修改。修改后下次刷新页面即会同步至
        {{ kungal.titleShort }}。
      </p>
    </div>

    <div class="flex justify-end">
      <KunButton :href="oauthProfileURL" target="_blank">
        前往 OAuth 账户中心
      </KunButton>
    </div>
  </KunCard>
</template>
