<script setup lang="ts">
import {
  GALGAME_RESOURCE_TYPE_ICON_MAP,
  GALGAME_RESOURCE_PLATFORM_ICON_MAP
} from '~/constants/galgameResource'
import {
  KUN_GALGAME_RESOURCE_TYPE_MAP,
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP
} from '~/constants/galgame'

const props = defineProps<{
  resource: GalgameResource
  refresh: () => void
}>()

const open = defineModel<boolean>({ required: true })

// Capture the Nuxt app at setup; reused by every post-await branch
// (handleReportExpire / handleDelete / handleEdit) to re-enter the
// captured Nuxt context. After `await kunFetch` resumes the active
// app instance is lost, so anything inside that touches
// useRuntimeConfig / useState / useFetch().refresh() crashes with
// "Cannot read properties of null (reading '$nuxt')" without
// runWithContext.
const nuxtApp = useNuxtApp()

const { id: currentUserId, role: currentUserRole } = usePersistUserStore()

// Local edit-modal state. Deliberately NOT going through
// useTempGalgameResourceStore + Resource.vue's KunModal +
// GalgameResourcePublish anymore — that triple-hop emit/store chain
// was where `$nuxt of null` kept resurfacing on edit-modal close (the
// refresh hop crossed too many post-await microtasks). The new
// LinkEditModal is fully local: own form ref, own kunFetch PUT, own
// refresh callback. See LinkEditModal.vue's header comment.
const isEditOpen = ref(false)

// Resource link/code/password are deliberately NOT in the summary
// payload — they're only fetched on demand to keep the list endpoint
// lightweight and avoid leaking links into search engines (the list
// API caches aggressively). Modal lazily fetches when opened the first
// time; subsequent re-opens reuse the cached detail.
const detail = ref<null | GalgameResourceDetailLink>(null)
const isFetching = ref(false)
const isExpired = computed(() => props.resource.status === 1)
const isOwner = computed(() => currentUserId === props.resource.user.id)
// Admin/moderator (role > 1) can edit / delete any resource — mirrors
// the rules on the dedicated detail page (resource/detail/Panel.vue).
const canManage = computed(() => isOwner.value || currentUserRole > 1)

const providerName = computed(() => {
  const names = props.resource.providerNames
  return names && names.length > 0
    ? names.join(' / ')
    : props.resource.linkDomain
})

const fetchDetail = async () => {
  if (detail.value || isFetching.value) return detail.value
  isFetching.value = true
  const result = await kunFetch<GalgameResourceDetailLink>(
    `/galgame-resource/${props.resource.id}/detail`,
    {
      method: 'GET',
      query: { galgameResourceId: props.resource.id }
    }
  )
  isFetching.value = false
  if (result) detail.value = result
  return detail.value
}

// Exposed so the parent (Link.vue) can run the fetch BEFORE flipping
// the modal open. Running fetch in the click handler's call stack keeps
// the Nuxt app context alive (versus firing from `watch(open)`, which
// runs in Vue's scheduler microtask where tryUseNuxtApp() returns null
// and kunFetch's first useRuntimeConfig crashes). The parent also gets
// to drive the button loading state directly off the returned promise.
defineExpose({ prefetch: fetchDetail })

// IMPORTANT: every kunFetch call below runs AFTER an `await` on the
// confirm alert (the user might sit on the dialog for many seconds),
// which loses the active Nuxt app context. Without runWithContext the
// `useRuntimeConfig` at the top of kunFetch hits `$nuxt of null`,
// kunFetch's catch returns null, the `if (result)` branch is skipped,
// and the button silently does nothing — exactly the "似乎失效" symptom.
const handleReportExpire = async () => {
  if (!currentUserId) {
    useMessage(10546, 'warn')
    return
  }
  const res = await useComponentMessageStore().alert(
    '您确定报告资源链接失效吗？',
    '这将通知资源发布者链接失效, 并将该链接标记为失效。若 17 天内资源发布者没有更换有效链接, 该资源链接将会被删除。恶意报告失效将会被处罚。'
  )
  if (!res) return

  isFetching.value = true
  const result = await nuxtApp.runWithContext(() =>
    kunFetch(`/galgame/${props.resource.galgameId}/resource/expired`, {
      method: 'PUT',
      body: { galgameResourceId: props.resource.id }
    })
  )
  isFetching.value = false

  if (result) {
    nuxtApp.runWithContext(() => {
      useMessage(10547, 'success')
      props.refresh()
      open.value = false
    })
  }
}

const handleDelete = async () => {
  const res = await useComponentMessageStore().alert(
    '您确定删除 Galgame 资源链接吗？',
    '这将扣除发布者获得的 5 萌萌点, 并扣除其它人对资源链接的点赞影响, 此操作不可撤销。'
  )
  if (!res) return

  isFetching.value = true
  const result = await nuxtApp.runWithContext(() =>
    kunFetch(`/galgame/${props.resource.galgameId}/resource`, {
      method: 'DELETE',
      query: { galgameResourceId: props.resource.id }
    })
  )
  isFetching.value = false

  if (result) {
    nuxtApp.runWithContext(() => {
      useMessage('删除资源成功', 'success')
      props.refresh()
      open.value = false
    })
  }
}

// Edit: simply flip a local ref. detail has been fetched on modal open
// (Link.vue's openDetail awaits prefetch first), so detail.value is
// guaranteed non-null by the time the user sees the 编辑 button.
const handleEdit = () => {
  if (!detail.value) return
  isEditOpen.value = true
}

// Called by LinkEditModal after a successful save: refresh the parent
// resource list AND dismiss the detail modal so the user returns to a
// fresh list view. Local detail.value is also nulled so the next
// 获取资源 click re-fetches (otherwise the modal would show stale
// values).
const handleEditDone = () => {
  detail.value = null
  props.refresh()
  open.value = false
}
</script>

<template>
  <KunModal v-model="open" inner-class-name="max-w-2xl w-[92vw] !p-0">
    <div class="flex flex-col">
      <!-- Header strip — colored by validity for at-a-glance signal -->
      <div
        :class="
          cn(
            'flex items-center justify-between gap-3 px-5 py-3',
            isExpired
              ? 'bg-warning/10 text-warning-700 dark:text-warning'
              : 'bg-success/10 text-success-700 dark:text-success'
          )
        "
      >
        <div class="flex items-center gap-2">
          <KunIcon
            :name="isExpired ? 'lucide:triangle-alert' : 'lucide:circle-check'"
            class="text-xl"
          />
          <span class="text-base font-medium">
            {{ isExpired ? '该资源链接已被标记失效' : '该资源链接可用' }}
          </span>
        </div>
        <KunChip
          variant="flat"
          :color="isExpired ? 'warning' : 'success'"
          size="sm"
        >
          {{ providerName }}
        </KunChip>
      </div>

      <div class="space-y-5 p-5">
        <!-- Publisher row -->
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-3">
            <KunAvatar :user="resource.user" size="lg" />
            <div class="flex flex-col">
              <span class="font-medium">{{ resource.user.name }}</span>
              <span class="text-default-500 text-xs">
                发布于 {{ formatTimeDifference(resource.created) }}
              </span>
            </div>
          </div>
          <KunChip variant="flat" color="default" size="sm">
            <KunIcon name="lucide:download" />
            {{ resource.download }} 次下载
          </KunChip>
        </div>

        <!-- Meta chips -->
        <div class="flex flex-wrap items-center gap-2">
          <KunChip color="primary" variant="flat">
            <KunIcon :name="GALGAME_RESOURCE_TYPE_ICON_MAP[resource.type]" />
            {{ KUN_GALGAME_RESOURCE_TYPE_MAP[resource.type] }}
          </KunChip>
          <KunChip color="warning" variant="flat">
            <KunIcon name="lucide:database" />
            {{ resource.size }}
          </KunChip>
          <KunChip color="success" variant="flat">
            <KunIcon
              :name="GALGAME_RESOURCE_PLATFORM_ICON_MAP[resource.platform]"
            />
            {{ KUN_GALGAME_RESOURCE_PLATFORM_MAP[resource.platform] }}
          </KunChip>
          <KunChip color="secondary" variant="flat">
            <KunIcon name="lucide:globe" />
            {{ KUN_GALGAME_RESOURCE_LANGUAGE_MAP[resource.language] }}
          </KunChip>
        </div>

        <!--
          Note from publisher — info-colored (cyan) to stand apart from
          the download block below. Many users skip notes by reflex; the
          color shift + leading "请先阅读" makes it harder to miss when
          scanning. Sits ABOVE the download links so the reader hits it
          on the way down.
        -->
        <KunInfo
          v-if="resource.note"
          color="info"
          variant="flat"
          title="发布者备注 — 请先阅读"
        >
          <p class="text-default-700 text-sm whitespace-pre-line">
            {{ resource.note }}
          </p>
        </KunInfo>

        <!-- Loading -->
        <div v-if="isFetching" class="flex justify-center py-8">
          <KunLoading />
        </div>

        <!-- Detail block — links + codes -->
        <template v-else-if="detail">
          <!--
            Download links — primary-colored (blue) so the visual flow
            reads "yellow → blue → status-colored copy buttons", with
            three distinct color buckets the eye can lock onto.
          -->
          <KunInfo color="primary" variant="flat" title="下载链接">
            <div class="space-y-1.5">
              <div
                v-for="(kun, index) in detail.link"
                :key="index"
                class="flex items-start gap-2"
              >
                <KunIcon
                  name="lucide:external-link"
                  class="text-primary mt-1 shrink-0"
                />
                <KunLink
                  :to="kun"
                  target="_blank"
                  rel="noopener noreferrer"
                  size="sm"
                  class-name="break-all"
                >
                  {{ kun }}
                </KunLink>
              </div>
            </div>
          </KunInfo>

          <!--
            Codes — separate block, status-colored. Different shape
            (clickable copy buttons) reinforces "these are action chips,
            not text blocks".
          -->
          <div
            v-if="detail.code || detail.password"
            class="flex flex-wrap items-center gap-2"
          >
            <KunCopy
              v-if="detail.code"
              variant="solid"
              :color="isExpired ? 'warning' : 'success'"
              :name="`提取码 ${detail.code}`"
              :text="detail.code"
            />
            <KunCopy
              v-if="detail.password"
              variant="solid"
              :color="isExpired ? 'warning' : 'success'"
              :name="`解压码 ${detail.password}`"
              :text="detail.password"
            />
          </div>

          <KunInfo color="danger" variant="bordered" title="补票提示">
            <p class="text-sm">
              Galgame 厂商制作游戏不易, 很多厂商如今都在炒冷饭,
              可见经济并不宽裕。 如果条件允许, 请尽可能前往
              <KunLink
                size="sm"
                :to="`/galgame/${resource.galgameId}`"
                class-name="inline"
              >
                Galgame 详情
              </KunLink>
              中的制作商部分进行正版补票, 感谢您对 Galgame 业界做出的贡献。
            </p>
          </KunInfo>
        </template>

        <!--
          Footer hierarchy:
            • One anchor button (关闭, solid) — the only filled button
              so the eye has exactly ONE thing to land on as "main exit".
            • Everything else is variant="light" — ghost-button look with
              colored text + hover bg. Color carries semantic only
              (danger=destructive, warning=report) instead of competing
              for visual weight.
            • Left cluster = resource-level actions (auth-gated).
            • Right cluster = navigation; the divider keeps the close
              button isolated as the visual destination.
        -->
        <div class="flex flex-wrap items-center justify-between gap-1">
          <div class="flex flex-wrap items-center gap-1">
            <KunButton
              v-if="canManage"
              variant="light"
              color="default"
              @click="handleEdit"
            >
              <KunIcon name="lucide:pencil" />
              编辑
            </KunButton>
            <KunButton
              v-if="canManage"
              variant="light"
              color="danger"
              :loading="isFetching"
              @click="handleDelete"
            >
              <KunIcon name="lucide:trash-2" />
              删除
            </KunButton>
            <KunButton
              v-if="!isOwner && !isExpired"
              variant="light"
              color="warning"
              :loading="isFetching"
              @click="handleReportExpire"
            >
              <KunIcon name="lucide:triangle-alert" />
              报告失效
            </KunButton>
          </div>

          <div class="flex flex-wrap items-center gap-1">
            <KunButton
              variant="light"
              color="default"
              :href="`/galgame-resource/${resource.id}`"
            >
              <KunIcon name="lucide:external-link" />
              查看详情页
            </KunButton>
            <KunButton variant="solid" color="default" @click="open = false">
              关闭
            </KunButton>
          </div>
        </div>
      </div>
    </div>

    <!--
      Inline edit modal. detail is guaranteed to exist here (handleEdit
      guards on it). After a successful save the modal closes itself
      and runs handleEditDone — which both refreshes the parent list
      and dismisses THIS detail modal so the user lands back on a
      fresh list.
    -->
    <GalgameResourceLinkEditModal
      v-if="detail"
      v-model="isEditOpen"
      :galgame-id="resource.galgameId"
      :resource="detail"
      :refresh="handleEditDone"
    />
  </KunModal>
</template>
