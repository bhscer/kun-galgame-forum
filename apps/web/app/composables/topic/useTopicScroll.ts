// Smooth-scroll to a topic reply (by floor) or comment (by id) and flash a brief
// highlight — the shared deep-link landing behaviour. Anchors:
//   - reply wrapper: id="<floor>.<slug>" (Reply.vue) → matched by the `<floor>.`
//     prefix, exact per floor (10. ≠ 1.).
//   - comment row:   id="comment-<id>"   (Comment.vue).
// Each returns whether the element was found, so the caller can surface a "target
// may be deleted / off-page" hint instead of silently doing nothing. Client-only
// (uses document); call from onMounted.
const FLASH = ['outline-2', 'outline-offset-2', 'outline-primary', 'rounded-lg']

const flash = (el: HTMLElement) => {
  el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  el.classList.add(...FLASH)
  setTimeout(() => el.classList.remove(...FLASH), 1500)
}

export const useTopicScroll = () => {
  const scrollToFloor = (floor: number): boolean => {
    if (!import.meta.client || floor <= 0) return false
    const el = document.querySelector<HTMLElement>(`[id^="${floor}."]`)
    if (!el) return false
    flash(el)
    return true
  }

  const scrollToComment = (commentId: number): boolean => {
    if (!import.meta.client || commentId <= 0) return false
    const el = document.getElementById(`comment-${commentId}`)
    if (!el) return false
    flash(el)
    return true
  }

  return { scrollToFloor, scrollToComment }
}
