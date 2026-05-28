import type { Ref } from 'vue'

interface ImageItem {
  src: string
  alt?: string
}

interface UseKunLightboxOptions {
  // Container to scan. Omitted → document-wide (page banners / heroes).
  // Provided → only the images inside that element register (one logical
  // gallery per block of content). KunContent passes its own
  // <article> ref so each prose block owns just its own images.
  scope?: Ref<HTMLElement | null>
  // Reactive getter to re-scan on. KunContent passes `() => props.content`
  // so a markdown swap (v-html replaces innerHTML) rebinds against the
  // fresh <img> nodes. Omit for static content that never changes after
  // mount (the document-wide banner case).
  watchSource?: () => unknown
}

// Wires every `data-kun-lazy-image` <img> in scope to open a KunLightbox.
// Single source of truth for the scan → map → bind → cleanup cycle so
// neither KunContent nor page-level consumers reimplement it.
export const useKunLightbox = (options: UseKunLightboxOptions = {}) => {
  const { scope, watchSource } = options
  const images = ref<ImageItem[]>([])
  const isLightboxOpen = ref(false)
  const currentImageIndex = ref(0)

  // Track bound elements so a content swap doesn't leak listeners. Old
  // <img> nodes are destroyed on swap (their listeners die with them),
  // but we clear refs explicitly and rebind against the new nodes.
  let bound: { el: HTMLImageElement; handler: () => void }[] = []

  const teardown = () => {
    for (const { el, handler } of bound) {
      el.removeEventListener('click', handler)
    }
    bound = []
  }

  const initializeLightbox = () => {
    teardown()
    const root: ParentNode | null = scope ? scope.value : document
    if (!root) {
      images.value = []
      return
    }

    let lazyImages = Array.from(
      root.querySelectorAll<HTMLImageElement>('img[data-kun-lazy-image]')
    )
    // Document-wide scan targets standalone images (banners, hero
    // covers). Markdown images live inside `<article class="kun-prose">`
    // and are owned by KunContent's own scoped instance — binding them
    // here too would race two lightboxes open on a single click.
    if (!scope) {
      lazyImages = lazyImages.filter((img) => !img.closest('.kun-prose'))
    }

    images.value = lazyImages.map((img) => ({
      src: img.getAttribute('src') || '',
      alt: img.getAttribute('alt') || ''
    }))

    lazyImages.forEach((img, index) => {
      img.style.cursor = 'zoom-in'
      const handler = () => {
        currentImageIndex.value = index
        isLightboxOpen.value = true
      }
      img.addEventListener('click', handler)
      bound.push({ el: img, handler })
    })
  }

  // nextTick so v-html / SSR-hydrated DOM is settled before we query it.
  onMounted(() => nextTick(initializeLightbox))
  onBeforeUnmount(teardown)

  if (watchSource) {
    watch(watchSource, () => nextTick(initializeLightbox), { flush: 'post' })
  }

  return {
    images,
    isLightboxOpen,
    currentImageIndex,
    initializeLightbox
  }
}
