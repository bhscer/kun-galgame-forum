<script setup lang="ts">
// Emitted whenever a menu item is activated so the parent popover (Avatar.vue)
// can dismiss itself — KunPopover stays open on inside-clicks by design.
const emit = defineEmits<{ close: [] }>()

const { id, name, moemoepoint, role, isCheckIn } = storeToRefs(
  usePersistUserStore()
)
const { messageStatus, showKUNGalgameMoemoepointLog } = storeToRefs(
  useTempSettingStore()
)

const isShowMessageDot = computed(() => messageStatus.value === 'new')
// role > 1 = 管理员 / 版主 (the /admin route is server-gated too; this just
// hides the entry from regular users, matching moyu's isModerator check).
const isAdmin = computed(() => role.value > 1)

// Opens the 萌萌点明细 modal, which is mounted at the stable Avatar.vue level
// (this menu lives inside a popover that unmounts on click-away).
const openMoemoepointLog = () => {
  emit('close')
  showKUNGalgameMoemoepointLog.value = true
}

const handleCheckIn = async () => {
  emit('close')
  isCheckIn.value = true

  const result = await kunFetch<number>('/user/check-in', {
    method: 'POST'
  })

  if (result === null) {
    return
  }

  moemoepoint.value += result

  if (result === 0) {
    useKunLoliInfo(
      '杂~~~鱼~♡杂鱼~♡ 臭杂鱼♡. 签到成功，您今日什么也没获得...',
      5000
    )
  } else if (result === 7) {
    useKunLoliInfo('杂鱼~♡♡♡♡♡. 签到成功, 您今日好运获得了 7 萌萌点哦!', 5000)
  } else {
    useKunLoliInfo(`杂~~~鱼~♡. 签到成功，您今日获得了 ${result} 萌萌点`, 5000)
  }
}

const logOut = async () => {
  emit('close')
  const res = await useComponentMessageStore().alert('您确定退出登录吗？')
  if (res) {
    useMessage(10110, 'success')
    usePersistUserStore().resetUser()
    // RP-initiated logout: clearing the forum's own state isn't enough — also
    // clear the central OP session, else the next login silently re-consents
    // into the same account. Top-level redirect to the OP logout entrypoint,
    // which returns to '/'. See docs/oauth/07-logout.md.
    startOAuthLogout()
  }
}
</script>

<template>
  <div class="flex flex-col gap-1">
    <div class="px-2 py-1">
      <p class="truncate font-semibold">{{ name }}</p>
    </div>

    <!-- 萌萌点 row doubles as the entry to the 明细 modal: "🍭 萌萌点 …… 8888 >"
         guides the user to click to view the full ledger. -->
    <button
      type="button"
      class="hover:bg-default-100 flex w-full items-center justify-between rounded-lg px-2 py-2 text-sm transition-colors"
      @click="openMoemoepointLog"
    >
      <span class="flex items-center gap-2">
        <KunIcon class="text-secondary size-4" name="lucide:lollipop" />
        萌萌点
      </span>
      <span class="flex items-center gap-1">
        <span class="text-secondary font-bold tabular-nums">
          {{ moemoepoint }}
        </span>
        <KunIcon class="text-foreground/40 size-4" name="lucide:chevron-right" />
      </span>
    </button>

    <NuxtLink
      :to="`/user/${id}/info`"
      class="hover:bg-default-100 flex items-center gap-2 rounded-lg px-2 py-2 text-sm transition-colors"
      @click="emit('close')"
    >
      <KunIcon class="size-4" name="lucide:user-round" />
      个人主页
    </NuxtLink>

    <NuxtLink
      to="/message"
      class="hover:bg-default-100 flex items-center gap-2 rounded-lg px-2 py-2 text-sm transition-colors"
      @click="emit('close')"
    >
      <KunIcon class="size-4" name="lucide:mail" />
      我的消息
      <span
        v-if="isShowMessageDot"
        class="bg-secondary-500 ml-auto size-2 rounded-full"
      />
    </NuxtLink>

    <NuxtLink
      v-if="isAdmin"
      to="/admin/overview"
      class="hover:bg-default-100 flex items-center gap-2 rounded-lg px-2 py-2 text-sm transition-colors"
      @click="emit('close')"
    >
      <KunIcon class="size-4" name="lucide:shield-check" />
      管理系统
    </NuxtLink>

    <KunButton
      v-if="!isCheckIn"
      variant="light"
      color="secondary"
      size="sm"
      :full-width="true"
      rounded="md"
      class-name="justify-between"
      @click="handleCheckIn"
    >
      <span class="flex items-center gap-2">
        <KunIcon class="size-4" name="lucide:calendar-check" />
        每日签到
      </span>
      <KunIcon class="text-secondary-500 size-5" name="lucide:sparkles" />
    </KunButton>

    <KunButton
      variant="light"
      color="danger"
      size="sm"
      :full-width="true"
      rounded="md"
      class-name="justify-start"
      @click="logOut"
    >
      <KunIcon class="size-4" name="lucide:log-out" />
      退出登录
    </KunButton>
  </div>
</template>
