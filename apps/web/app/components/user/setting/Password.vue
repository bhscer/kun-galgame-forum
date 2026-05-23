<script setup lang="ts">
// Password change moved to OAuth profile per the 2026-05-23 policy
// (docs/oauth/README.md §身份操作必须在 OAuth 完成 + docs/oauth/
// 02-user-profile.md §身份操作 vs 展示操作). Like email, the in-app
// change-password page is gone; users jump to OAuth account-center.
//
// Same OAuth-base resolution as Email.vue / legacy Avatar: strip
// /api/vN/ off the configured OAuth server URL to land on /profile.
const oauthProfileURL = computed(() => {
  const apiBase = useRuntimeConfig().public.oauthServerUrl || ''
  return apiBase.replace(/\/api\/v\d+\/?$/, '') + '/profile'
})
</script>

<template>
  <KunCard :is-hoverable="false" content-class="space-y-3">
    <div class="space-y-2">
      <span class="text-xl">更改密码</span>
      <p class="text-default-500 text-sm">
        密码由 OAuth 账户中心统一管理。修改密码需要验证旧密码,
        请前往 OAuth 账户中心进行修改。忘记密码同样在那里处理。
      </p>
    </div>

    <div class="flex justify-end">
      <KunButton :href="oauthProfileURL" target="_blank">
        前往 OAuth 账户中心
      </KunButton>
    </div>
  </KunCard>
</template>
