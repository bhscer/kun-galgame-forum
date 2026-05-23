<script setup lang="ts">
defineProps<{
  user: UserInfo
}>()
</script>

<template>
  <div class="h-full w-full space-y-6">
    <KunHeader
      name="信息设置"
      description="您可以更改您的头像, 签名, 用户名。邮箱和密码在 OAuth 账户中心管理。"
    />

    <UserSettingAvatar />
    <UserSettingBio :user="user" />
    <UserSettingUsername />
    <UserSettingEmail />
    <UserSettingPassword />
  </div>
</template>
