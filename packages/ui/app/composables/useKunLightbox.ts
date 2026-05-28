import type { Ref } from 'vue'

interface ImageItem {
  src: string
  alt?: string
}

interface UseKunLightboxOptions {
  // Container whose `data-kun-lazy-image` <img>s register as one gallery.
  // One logical gallery per block (a reply's screenshots stay separate
  // from another reply's).
  scope: Ref<HTMLElement | null>
  // Reactive getter to re-scan on (e.g. `() => props.content`) so a
  // v-html swap rebinds against the fresh <img> nodes. Omit when the
  // content never changes after mount.
  watchSource?: () => unknown
}

// Wires every `data-kun-lazy-image` <img> inside `scope` to open a
// KunLightbox. The single source of truth for the scan → map → bind →
// cleanup cycle so consumers (KunContent et al.) don't reimplement it.
export const useKunLightbox = (options: UseKunLightboxOptions) => {
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
    const root = scope.value
    if (!root) {
      images.value = []
      return
    }

    const lazyImages = Array.from(
      root.querySelectorAll<HTMLImageElement>('img[data-kun-lazy-image]')
    )
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
    currentImageIndex
  }
}
