<script setup lang="ts">
// GALGAME_CREATION — "new galgame" feed card (layout per spec):
//   top    — avatar · username · time (shell), then "创建了一个新的Galgame,已经有
//            N 个下载资源"
//   middle — banner (left) + name · 发售于<date> (right)
//   bottom — 点赞 + 收藏 buttons (reused detail components) + 查看详情 link
// 制作会社 + 简介 are intentionally omitted until the wiki brief exposes them.
const props = defineProps<{ activity: ActivityItem }>()

const data = computed(
  () => props.activity.data as GalgameActivityData | undefined
)
const gid = computed(() => data.value?.galgameId ?? 0)
const detailLink = computed(() =>
  gid.value ? `/galgame/${gid.value}` : props.activity.link
)

// 由 <制作会社> 制作，发售于 <发售日期> — each half shown only if present.
const madeBy = computed(() => {
  const parts: string[] = []
  if (data.value?.developer) parts.push(`由 ${data.value.developer} 制作`)
  if (data.value?.releaseDate) parts.push(`发售于 ${data.value.releaseDate}`)
  return parts.join('，')
})

// Hydrate the like/favorite buttons' initial state (the cached feed can't carry
// it); fetched once per session, client-side.
const { isLiked, isFavorited, ensureLoaded } = useMyGalgameInteractions()
onMounted(ensureLoaded)
</script>

<template>
  <ActivityCardShell :actor="activity.actor" :timestamp="activity.timestamp">
    <div class="space-y-3">
      <p class="text-default-600 text-sm">
        创建了一个新的 Galgame,已经有 {{ data?.resourceCount ?? 0 }} 个下载资源
      </p>

      <div class="flex items-start gap-3">
        <KunLink :to="detailLink" class-name="shrink-0">
          <div
            class="bg-default-100 aspect-video w-32 overflow-hidden rounded-lg sm:w-44"
          >
            <img
              v-if="data?.coverHash"
              :src="imageHashUrl(imageCdnBase(), data.coverHash, 'mini')"
              :alt="data?.name"
              loading="lazy"
              class="h-full w-full object-cover"
            />
          </div>
        </KunLink>

        <div class="min-w-0 flex-1 space-y-1">
          <KunLink
            underline="none"
            color="default"
            :to="detailLink"
            class-name="hover:text-primary block"
          >
            <h3 class="line-clamp-2 font-medium break-all">
              {{ data?.name || activity.content }}
            </h3>
          </KunLink>
          <p v-if="madeBy" class="text-default-700 text-sm break-all">
            {{ madeBy }}
          </p>
          <p
            v-if="data?.intro"
            class="text-default-500 line-clamp-3 text-sm break-all"
          >
            {{ markdownToText(data.intro) }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <GalgameLike
          :galgame-id="gid"
          :target-user-id="activity.actor?.id ?? 0"
          :like-count="data?.likeCount ?? 0"
          :is-liked="isLiked(gid)"
        />
        <GalgameFavorite
          :galgame-id="gid"
          :target-user-id="activity.actor?.id ?? 0"
          :favorite-count="data?.favoriteCount ?? 0"
          :is-favorited="isFavorited(gid)"
        />
        <KunLink
          underline="none"
          color="default"
          :to="detailLink"
          class-name="text-default-500 hover:text-primary ml-auto flex items-center gap-0.5 text-sm"
        >
          查看详情
          <KunIcon name="lucide:chevron-right" class="size-4" />
        </KunLink>
      </div>
    </div>
  </ActivityCardShell>
</template>
