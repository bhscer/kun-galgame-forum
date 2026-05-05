<script setup lang="ts">
const oauthProfileURL = computed(() => {
  // oauthServerUrl 是 .../api/v1 形式；剥到根域作为账户中心入口
  const apiBase = useRuntimeConfig().public.oauthServerUrl || ''
  return apiBase.replace(/\/api\/v\d+\/?$/, '') + '/profile'
})
</script>

<template>
  <KunCard :is-hoverable="false" content-class="space-y-3">
    <div class="space-y-2">
      <span class="text-xl">更改头像</span>
      <p class="text-default-500 text-sm">
        头像现在统一由 OAuth 账户中心管理。请前往 OAuth
        账户中心修改，修改后下次刷新页面即会同步至 {{ kungal.titleShort }}。
      </p>
      <p class="text-default-500 text-sm">
        您的默认头像将会从
        <KunLink size="sm" :to="kungal.domain.sticker" target="_blank">
          鲲 Galgame 表情包
        </KunLink>
        中随机选取, 每一次都是不同的孩子哦, 欸嘿嘿嘿
      </p>
    </div>

    <div class="flex justify-end">
      <KunButton :href="oauthProfileURL" target="_blank">
        前往 OAuth 账户中心
      </KunButton>
    </div>
  </KunCard>
</template>
