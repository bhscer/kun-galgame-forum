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

// Logout opens a scope chooser: "this site only" vs "this site + OAuth". Two
// session layers exist (forum's own + the central OAuth/SSO session); we let the
// user pick which to end and explain the impact. See docs/oauth/07-logout.md.
const showLogoutModal = ref(false)
const logoutPending = ref<'local' | 'everywhere' | null>(null)

const logOut = () => {
  emit('close')
  showLogoutModal.value = true
}

// "This site only" — reset the forum's own session; the OAuth (SSO) session and
// other sites stay logged in, so re-login here is silent (auto-consent).
const logoutLocal = () => {
  if (logoutPending.value) return
  logoutPending.value = 'local'
  usePersistUserStore().resetUser()
  useMessage(10110, 'success')
  showLogoutModal.value = false
  logoutPending.value = null
}

// "Everywhere" — also end the central OP session via RP-initiated logout, so no
// site can silently re-login and the OAuth account is signed out.
const logoutEverywhere = () => {
  if (logoutPending.value) return
  logoutPending.value = 'everywhere'
  usePersistUserStore().resetUser()
  startOAuthLogout() // top-level redirect to the OP; clears the SSO session
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

    <KunModal v-model="showLogoutModal" inner-class-name="max-w-lg">
      <div class="space-y-4 p-2">
        <div class="space-y-1">
          <h2 class="text-foreground text-lg font-semibold">退出登录</h2>
          <p class="text-default-500 text-sm">请选择退出范围：</p>
        </div>

        <button
          type="button"
          :disabled="!!logoutPending"
          class="border-primary-200 bg-primary-50/50 hover:bg-primary-100/60 w-full rounded-xl border p-4 text-left transition-colors disabled:opacity-60"
          @click="logoutEverywhere"
        >
          <div class="flex items-start gap-3">
            <KunIcon
              :name="logoutPending === 'everywhere' ? 'lucide:loader-circle' : 'lucide:log-out'"
              :class="`text-primary mt-0.5 size-5 shrink-0 ${logoutPending === 'everywhere' ? 'animate-spin' : ''}`"
            />
            <div class="space-y-1">
              <div class="text-foreground flex items-center gap-2 font-medium">
                退出本站和 OAuth 账号
                <span class="bg-primary-100 text-primary-700 rounded px-1.5 py-0.5 text-xs">推荐</span>
              </div>
              <p class="text-default-500 text-xs leading-relaxed">
                本站与 OAuth 账号都会退出；其它已登录的站点会在下次刷新登录态时一并退出；再次登录需重新验证身份。适合公共 / 共享设备。
              </p>
            </div>
          </div>
        </button>

        <button
          type="button"
          :disabled="!!logoutPending"
          class="border-default-200 hover:bg-default-100 w-full rounded-xl border p-4 text-left transition-colors disabled:opacity-60"
          @click="logoutLocal"
        >
          <div class="flex items-start gap-3">
            <KunIcon
              :name="logoutPending === 'local' ? 'lucide:loader-circle' : 'lucide:monitor'"
              :class="`text-default-500 mt-0.5 size-5 shrink-0 ${logoutPending === 'local' ? 'animate-spin' : ''}`"
            />
            <div class="space-y-1">
              <div class="text-foreground font-medium">仅退出本站</div>
              <p class="text-default-500 text-xs leading-relaxed">
                只退出本站；OAuth 账号与其它站点保持登录；再次登录本站可免密直接进入。适合自己的设备。
              </p>
            </div>
          </div>
        </button>

        <div class="flex justify-end pt-1">
          <KunButton
            variant="light"
            :disabled="!!logoutPending"
            @click="showLogoutModal = false"
          >
            取消
          </KunButton>
        </div>
      </div>
    </KunModal>
  </div>
</template>
