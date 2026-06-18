<script setup lang="ts">
interface CreatorEligibility {
  eligible: boolean
  merged_prs: number
  galgames_published: number
  reviews_100: number
  need_merged_prs: number
  need_galgames: number
  need_reviews: number
}

interface CreatorApplicationInfo {
  id: number
  status: string
  decline_reason: string
}

interface CreatorStatus {
  eligibility: CreatorEligibility
  application: CreatorApplicationInfo | null
  // Whether the user already holds the creator role (source of truth — covers
  // an admin-granted role with no approved application, and the post-approval
  // window before the role cache refreshes).
  is_creator: boolean
}

const STATUS_LABEL: Record<string, string> = {
  pending: '审核中',
  approved: '已通过',
  declined: '已拒绝'
}
const STATUS_COLOR: Record<string, 'primary' | 'success' | 'warning'> = {
  pending: 'primary',
  approved: 'success',
  declined: 'warning'
}

const status = ref<CreatorStatus | null>(null)
const message = ref('')
const loading = ref(true)
const failed = ref(false)
const submitting = ref(false)

const load = async () => {
  loading.value = true
  failed.value = false
  const res = await kunFetch<CreatorStatus>('/user/creator/status')
  if (res) {
    status.value = res
  } else if (!status.value) {
    // kunFetch already toasts the error; show an inline retry only when we have
    // nothing to display (a failed refresh keeps the last good status visible).
    failed.value = true
  }
  loading.value = false
}

// Already a creator — terminal success state; no progress/apply UI to show.
// Prefer the actual role (is_creator); fall back to an approved application for
// the brief window where the role grant hasn't propagated to the role cache.
const isCreator = computed(
  () =>
    !!status.value?.is_creator ||
    status.value?.application?.status === 'approved'
)

// A pending application blocks re-application; declined/none allows (re)apply.
const canApply = computed(
  () =>
    !!status.value?.eligibility.eligible &&
    status.value?.application?.status !== 'pending'
)

const handleApply = async () => {
  submitting.value = true
  const res = await kunFetch<CreatorApplicationInfo>('/user/creator/apply', {
    method: 'POST',
    body: { message: message.value }
  })
  submitting.value = false
  if (res) {
    useMessage('申请已提交，等待管理员审核', 'success')
    message.value = ''
    await load()
  }
}

onMounted(load)
</script>

<template>
  <KunCard :is-hoverable="false" content-class="space-y-4">
    <div>
      <span class="text-xl">创作者申请</span>
      <p class="text-default-500 text-sm">
        创作者可直接发布 Galgame 词条 (含无 VNDB ID 的同人 /
        独立作品)。满足以下任一条件即可申请，提交后由管理员审核。
      </p>
    </div>

    <template v-if="status">
      <!-- Already a creator: terminal success state, no progress/apply UI. -->
      <div v-if="isCreator" class="space-y-1">
        <KunChip color="success">🎉 你已是创作者</KunChip>
        <p class="text-default-500 text-sm">
          你已拥有创作者权限，可直接发布 Galgame 词条，无需再次申请。
        </p>
      </div>

      <template v-else>
        <div class="space-y-2">
          <div class="flex items-center justify-between text-sm">
            <span>合并的 PR</span>
            <span
              :class="
                status.eligibility.merged_prs >=
                status.eligibility.need_merged_prs
                  ? 'text-success'
                  : 'text-default-500'
              "
            >
              {{ status.eligibility.merged_prs }} /
              {{ status.eligibility.need_merged_prs }}
            </span>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span>已发布 Galgame</span>
            <span
              :class="
                status.eligibility.galgames_published >=
                status.eligibility.need_galgames
                  ? 'text-success'
                  : 'text-default-500'
              "
            >
              {{ status.eligibility.galgames_published }} /
              {{ status.eligibility.need_galgames }}
            </span>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span>百字以上简评 (≥100 字)</span>
            <span
              :class="
                status.eligibility.reviews_100 >=
                status.eligibility.need_reviews
                  ? 'text-success'
                  : 'text-default-500'
              "
            >
              {{ status.eligibility.reviews_100 }} /
              {{ status.eligibility.need_reviews }}
            </span>
          </div>
        </div>

        <div v-if="status.application" class="flex items-center gap-2 text-sm">
          <span>当前申请：</span>
          <KunChip
            :color="STATUS_COLOR[status.application.status] || 'default'"
          >
            {{
              STATUS_LABEL[status.application.status] ||
              status.application.status
            }}
          </KunChip>
          <span
            v-if="
              status.application.status === 'declined' &&
              status.application.decline_reason
            "
            class="text-default-500"
          >
            原因：{{ status.application.decline_reason }}
          </span>
        </div>

        <template v-if="status.application?.status !== 'pending'">
          <KunChip :color="status.eligibility.eligible ? 'success' : 'warning'">
            {{
              status.eligibility.eligible ? '符合申请条件' : '尚不符合申请条件'
            }}
          </KunChip>

          <KunTextarea
            v-if="status.eligibility.eligible"
            name="creator-message"
            placeholder="(可选) 附言：向管理员说明你的情况"
            :rows="3"
            v-model="message"
          />

          <div class="flex justify-end">
            <KunButton
              :disabled="!canApply"
              :loading="submitting"
              @click="handleApply"
            >
              {{
                status.application?.status === 'declined'
                  ? '重新申请'
                  : '提交申请'
              }}
            </KunButton>
          </div>
        </template>
      </template>
    </template>

    <p v-else-if="loading" class="text-default-500 text-sm">加载中...</p>

    <div v-else-if="failed" class="flex items-center justify-between text-sm">
      <span class="text-default-500">加载失败</span>
      <KunButton size="sm" variant="flat" @click="load">重试</KunButton>
    </div>
  </KunCard>
</template>
