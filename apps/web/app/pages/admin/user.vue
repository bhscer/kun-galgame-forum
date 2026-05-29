<script setup lang="ts">
import { watchDebounced } from '@vueuse/core'

useKunDisableSeo('用户内容管理')

// Account details (ban / delete / profile) live in the OAuth admin UI now.
// This page only manages the CONTENT a user has published on kungal — find
// a user, then purge everything they posted here (for spam / ad accounts).
const searchQuery = ref('')
const users = ref<SearchResultUser[]>([])
const isSearching = ref(false)

const handleSearch = async () => {
  const keywords = searchQuery.value.trim()
  if (!keywords) {
    users.value = []
    return
  }
  isSearching.value = true
  const res = await kunFetch<{ items: SearchResultUser[]; total: number }>(
    '/search',
    { method: 'GET', query: { keywords, type: 'user', page: 1, limit: 30 } }
  )
  isSearching.value = false
  users.value = res?.items ?? []
}

watchDebounced(() => searchQuery.value, handleSearch, {
  debounce: 500,
  maxWait: 1000
})
</script>

<template>
  <KunCard :is-hoverable="false" :is-transparent="false">
    <KunHeader
      name="用户内容管理"
      description="账号本身的封禁 / 删除已迁移至 OAuth 管理后台。此处用于管理用户在本站发布的内容: 搜索用户后, 可一键清除其在 kungal 的全部内容 (话题 / 回复 / 评论 / 评分 / 资源 / 网站 / 工具及一切互动), 主要用于清理广告与 spam 账号。单条内容的编辑与删除请直接在对应页面操作。"
    >
      <template #endContent>
        <KunInput
          v-model="searchQuery"
          type="text"
          placeholder="输入用户名以搜索用户"
        />
      </template>
    </KunHeader>

    <div class="mt-6 flex flex-col gap-3">
      <AdminUserCard v-for="user in users" :key="user.id" :user="user" />
    </div>

    <KunLoading v-if="isSearching" />
    <KunNull
      v-if="!isSearching && !users.length && searchQuery.trim()"
      description="未找到匹配的用户"
    />
  </KunCard>
</template>
