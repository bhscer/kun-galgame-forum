<script setup lang="ts">
// Reaction chips below a topic / reply: reactor avatars when count < 5, else
// emoji + count; tap a chip to toggle. State is provided by the parent
// (useReactions) so these chips and the trigger button stay in sync even when
// the trigger lives elsewhere (e.g. the desktop footer).
import { KUN_REACTION_EMOJI, reactionAsset } from '~/constants/reaction'

const { list, toggle } = inject(reactionsKey)!

const AVATAR_THRESHOLD = 5
const showAvatars = (r: KunReaction) =>
  r.count < AVATAR_THRESHOLD && !!r.reactors?.length
</script>

<template>
  <div v-if="list.length" class="flex flex-wrap items-center gap-1.5">
    <button
      v-for="r in list"
      :key="r.reaction"
      type="button"
      :class="
        cn(
          'flex items-center gap-1 rounded-full border px-2 py-0.5 text-sm transition-colors',
          r.mine
            ? 'border-primary bg-primary/10 text-primary'
            : 'border-default-200 text-default-600 hover:bg-default-100'
        )
      "
      @click="toggle(r.reaction)"
    >
      <img
        :src="reactionAsset(r.reaction)"
        :alt="KUN_REACTION_EMOJI[r.reaction] ?? r.reaction"
        class="size-5 shrink-0 max-w-none"
        loading="lazy"
      />
      <span v-if="showAvatars(r)" class="flex -space-x-1.5">
        <KunAvatar
          v-for="u in r.reactors!.slice(0, 4)"
          :key="u.id"
          :user="u"
          size="sm"
          :is-navigation="false"
        />
      </span>
      <span v-else class="tabular-nums">{{ formatNumber(r.count) }}</span>
    </button>
  </div>
</template>
