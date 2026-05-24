<script setup lang="ts">
// Renders the OAuth-jump register button. Pre-L1 this page was a 301
// to /login because no register UX existed; L1 restored an actual
// register flow via OAuth's hosted register page (see
// components/register/Register.vue + docs/oauth/05-registration.md).
//
// The top-bar's "注册" button (components/kun/top-bar/Avatar.vue:76)
// points here, so the route must stay live.
useKunSeoMeta({
  title: '注册',
  description: `注册 ${kungal.titleShort}`
})
</script>

<template>
  <Register />
</template>
