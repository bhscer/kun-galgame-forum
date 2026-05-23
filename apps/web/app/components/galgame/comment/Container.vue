<script setup lang="ts">
const props = defineProps<{
  userData: KunUser[]
  targetUser: KunUser
}>()

const route = useRoute()
const { commentToUserId } = storeToRefs(useTempGalgameCommentStore())

const username = ref(props.targetUser.name)
const gid = parseInt((route.params as { gid: string }).gid)

const pageData = reactive({
  galgameId: gid,
  page: 1,
  limit: 30,
  sortOrder: 'desc'
})

// data.items now holds ROOTS only — replies are nested inside each
// root.replies (depth-unbounded subtree); the frontend caps visible
// recursion at depth 3 in Comment.vue and routes deeper exploration to
// ThreadDrawer. data.total counts roots so pagination math matches the
// list of visible cards.
const { data, status, refresh } = await useKunFetch(
  `/galgame/${gid}/comment/all`,
  {
    lazy: true,
    method: 'GET',
    query: pageData
  }
)

const handleSetUserInfo = (name: string) => {
  username.value = name
  commentToUserId.value =
    props.userData.find((user) => user.name === name)?.id ||
    props.targetUser.id
}

onMounted(() => (commentToUserId.value = props.targetUser.id))

// Drawer state: when a deep reply hits the recursion cap, the
// corresponding Comment node bubbles up an open-thread event with the
// thread's root id. Setting this back to null closes the drawer.
const openThreadRootId = ref<number | null>(null)
</script>

<template>
  <div class="space-y-3">
    <KunHeader name="游戏评论" scale="h2">
      <template #endContent>
        <KunLink size="sm" to="/topic/1482">
          Galgame 评论注意事项, 资源失效, 解压密码错误等问题反馈
        </KunLink>
      </template>
    </KunHeader>

    <div v-if="targetUser" class="flex items-center gap-2">
      <div class="whitespace-nowrap">评论给</div>
      <KunSelect
        :model-value="username"
        :options="
          userData.map((user) => ({ value: user.name, label: user.name }))
        "
        @set="(value) => handleSetUserInfo(value.toString())"
      >
        {{ username }}
      </KunSelect>
    </div>

    <div v-if="data" class="space-y-3">
      <GalgameCommentPanel :refresh="refresh">
        <div v-if="data.total" class="flex items-center gap-2">
          <KunButton
            :is-icon-only="true"
            :variant="pageData.sortOrder === 'desc' ? 'flat' : 'light'"
            size="lg"
            @click="pageData.sortOrder = 'desc'"
          >
            <KunIcon class="text-inherit" name="lucide:arrow-down" />
          </KunButton>

          <KunButton
            :is-icon-only="true"
            :variant="pageData.sortOrder === 'asc' ? 'flat' : 'light'"
            size="lg"
            @click="pageData.sortOrder = 'asc'"
          >
            <KunIcon class="text-inherit" name="lucide:arrow-up" />
          </KunButton>
        </div>
      </GalgameCommentPanel>

      <KunLoading v-if="status === 'pending'" />

      <KunNull
        v-if="!data.total && status !== 'pending'"
        description="没人评论, 是没人要这个 Galgame 的小只可爱软萌女孩子了吗, 呜呜呜呜呜呜！！"
      />

      <div v-if="status !== 'pending' && data.total" class="space-y-6">
        <GalgameComment
          v-for="comment in data.items"
          :key="comment.id"
          :comment="comment"
          :refresh="refresh"
          :depth="0"
          @open-thread="(rootId) => (openThreadRootId = rootId)"
        />
      </div>

      <KunPagination
        v-if="data.total > 30 || data.total === 30"
        v-model:current-page="pageData.page"
        :total-page="Math.ceil(data.total / pageData.limit)"
        :is-loading="status === 'pending'"
      />
    </div>

    <GalgameCommentThreadDrawer
      v-model:root-comment-id="openThreadRootId"
      :galgame-id="gid"
      :refresh="refresh"
    />
  </div>
</template>
