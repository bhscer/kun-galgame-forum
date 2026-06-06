import { JSDOM } from 'jsdom'
import createDOMPurify, { type Config } from 'dompurify'
import { __setKunServerSanitize } from '../utils/sanitize'

// Server-side sanitizer injected into utils/sanitize.ts. This file is
// `.server` so the static jsdom import is server-only (out of the client
// bundle) AND stays external/complete (a dynamic import inlined + broke jsdom).
//
// jsdom is the only server DOM that sanitizes correctly (linkedom / happy-dom
// silently fail to strip <script>/javascript:). But it retains a DOM graph per
// sanitize under SSR (heap snapshots showed SymbolTreeNode / nwsapi Finder /
// *Impl climbing ~100 MiB/min → OOM). Two mitigations:
//   - recycle + close() the jsdom window every RECYCLE_EVERY calls so the
//     accumulation is dropped + GC'd (bounded sawtooth, not unbounded growth);
//   - an LRU memo so identical content (a topic body + each reply, re-rendered
//     every visit) sanitizes once per unique input, not per render — cutting
//     jsdom churn + the sanitize CPU to ~zero on hot pages.
export default defineNuxtPlugin(() => {
  const RECYCLE_EVERY = 200
  const CACHE_MAX = 2000

  let dom = new JSDOM('')
  let purifier = createDOMPurify(dom.window as never)
  let calls = 0
  const cache = new Map<string, string>()

  const recycle = () => {
    const old = dom
    dom = new JSDOM('')
    purifier = createDOMPurify(dom.window as never)
    try {
      old.window.close()
    } catch {
      // best-effort; dropping the reference is enough for GC
    }
  }

  __setKunServerSanitize((dirty: string, config?: Config): string => {
    const key = (config ? JSON.stringify(config) : '') + ' ' + dirty
    const hit = cache.get(key)
    if (hit !== undefined) {
      cache.delete(key)
      cache.set(key, hit)
      return hit
    }

    if (++calls >= RECYCLE_EVERY) {
      recycle()
      calls = 0
    }
    const clean = purifier.sanitize(dirty, config) as unknown as string

    cache.set(key, clean)
    if (cache.size > CACHE_MAX) {
      const oldest = cache.keys().next().value
      if (oldest !== undefined) cache.delete(oldest)
    }
    return clean
  })
})
