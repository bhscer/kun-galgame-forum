<script setup lang="ts">
import DOMPurify from 'isomorphic-dompurify'
import { useSpoilerContent } from '../../../composables/topic/useSpoilerContent'
import 'katex/dist/katex.min.css'

const props = withDefaults(
  defineProps<{
    content: string
    className?: string
  }>(),
  {
    className: ''
  }
)

const articleRef = ref<HTMLElement | null>(null)

useSpoilerContent(articleRef)

const sanitizeConfig = {
  ADD_TAGS: ['div', 'span', 'button'],
  ADD_ATTR: ['class', 'title', 'line']
}

// Click-to-zoom for every `data-kun-lazy-image` <img> the backend
// renders into this prose block. KunContent is the single markdown
// surface site-wide, so wiring the lightbox here gives every consumer
// (topic body + replies, galgame intro, comments, toolset, doc
// article, user galgame comments…) zoomable images for free.
//
// scope = this article ref → one logical gallery per prose block (a
// reply's screenshots stay separate from another reply's). watchSource
// = content → rebind when the markdown swaps, which also covers
// async-loaded replies/comments that mount their own KunContent.
const {
  images: lightboxImages,
  isLightboxOpen,
  currentImageIndex: lightboxIndex
} = useKunLightbox({
  scope: articleRef,
  watchSource: () => props.content
})
</script>

<template>
  <div>
    <article
      ref="articleRef"
      :class="cn('kun-prose', className)"
      v-html="DOMPurify.sanitize(content, sanitizeConfig)"
    />

    <!-- Only mount the dialog when this prose actually has images.
         Topic pages render dozens of KunContent instances (every reply
         + comment); skipping the empty ones avoids dozens of idle
         <dialog> nodes. SSR + client-initial both see [] (the scan runs
         in onMounted, client-only), so there's no hydration mismatch. -->
    <KunLightbox
      v-if="lightboxImages.length"
      :images="lightboxImages"
      v-model:is-open="isLightboxOpen"
      :initial-index="lightboxIndex"
    />
  </div>
</template>

<style scoped>
.kun-prose {
  & :deep(.kun-spoiler) {
    position: relative;
    display: inline-block;
    border-radius: 0.5rem;
    overflow: hidden;
    vertical-align: middle;

    & .spoiler-frost {
      position: absolute;
      inset: 0;
      pointer-events: none;
      background: rgb(150, 150, 150);
      border-radius: inherit;
      transform: translateZ(0);
      z-index: 5;
    }

    & > *:not(.spoiler-canvas) {
      transition: filter 0.3s ease-in-out;
      filter: blur(0);
    }

    & .spoiler-canvas {
      position: absolute;
      inset: 0;
      z-index: 1;
      pointer-events: none;
      opacity: 1;
      transition: opacity 0.3s ease-in-out;
      background-color: rgba(150, 150, 150, 0.1);

      &.fade-out {
        opacity: 0;
      }
    }
  }

  & :deep(.kun-spoiler.kun-spoiler-hidden) {
    cursor: pointer;

    & > *:not(.spoiler-canvas) {
      filter: blur(52px);
      user-select: none;
    }
  }

  & :deep(div.kun-spoiler) {
    display: block;
    width: fit-content;
  }
}
</style>
