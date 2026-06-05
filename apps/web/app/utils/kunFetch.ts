/**
 * Unified API response format from Go backend.
 * All endpoints return: { code: 0, message: "成功", data: T }
 */
interface KunApiResponse<T> {
  code: number
  message: string
  data: T
}

// Hard ceiling for SSR API calls. Without it, a stalled server-side $fetch to
// the Go backend (a half-open keep-alive socket left over after an api
// redeploy, a transient network stall) never settles: the awaiting
// useAsyncData/Suspense render hangs forever, pinning its entire SSR render
// context in the heap and leaking the in/out sockets (incoming CLOSE_WAIT +
// outgoing ESTABLISHED to the api). Node never reaps a slow handler, so these
// accumulate over days until the event loop is GC-thrashing at 100% CPU.
// ofetch's `timeout` aborts via AbortController, so a stuck request fails fast
// and the handler unwinds. SSR-only on purpose: a hung client fetch can't leak
// the server, and client paths (large file uploads) legitimately run long.
const SSR_API_TIMEOUT_MS = 10000

const CODE_AUTH_EXPIRED = 205
// CODE_BANNED comes from kungal's mapping of OAuth 10014. We MUST NOT
// redirect banned users to /login — a fresh login would just hit 10014
// again at the next refresh, putting them in an infinite loop. Instead
// surface the message prominently and stop there.
const CODE_BANNED = 234

// Only forward the auth session cookie to the Go backend during SSR.
//
// pinia-plugin-persistedstate serializes every persisted Pinia store
// into its own browser cookie (user profile, settings panel, sidebar
// state, etc.). Accumulated, the full Cookie header easily exceeds
// Fiber's default ReadBufferSize (4KB) — manifests as
// `Request Header Fields Too Large` on Go and silently-empty SSR
// renders (the data fetch fails, the page hydrates with no payload).
//
// The Go backend only authenticates against `kungal_session`; the
// other cookies are pure client state and would be ignored anyway.
// Forwarding only the one keeps SSR auth working without dragging
// every persisted store across the wire.
const SESSION_COOKIE_NAME = 'kungal_session'

const extractSessionCookie = (
  cookieHeader?: string
): string | undefined => {
  if (!cookieHeader) return undefined
  const prefix = `${SESSION_COOKIE_NAME}=`
  for (const part of cookieHeader.split(';')) {
    const trimmed = part.trim()
    if (trimmed.startsWith(prefix)) return trimmed
  }
  return undefined
}

const handleApiError = async (code: number, message: string) => {
  if (import.meta.server) return

  if (code === CODE_BANNED) {
    const userStore = usePersistUserStore()
    if (userStore.id) {
      userStore.resetUser()
    }
    useMessage(message || '您的账号已被封禁', 'error', 10000)
    return
  }

  if (code === CODE_AUTH_EXPIRED) {
    const { default: Cookies } = await import('js-cookie')
    const navigateCookie = Cookies.get('kun-is-navigate-to-login')
    const userStore = usePersistUserStore()

    if (!navigateCookie && userStore.id) {
      userStore.resetUser()
      useMessage(message || '登录已失效，请重新登录', 'error', 7777)
      Cookies.set('kun-is-navigate-to-login', 'navigated', {
        expires: 1 / 1440
      })
      await navigateTo('/')
    }
    return
  }

  if (code !== 0) {
    useMessage(message, 'error')
  }
}

/**
 * useKunFetch — SSR-safe composable built on Nuxt 4 `createUseFetch`.
 *
 * Automatically:
 * - Resolves baseURL for SSR/CSR
 * - Forwards cookies during SSR via credentials
 * - Unwraps `{ code, data }` response
 * - Handles auth/biz errors client-side only
 *
 * The response type T is what the Go backend returns inside `data`.
 * The transform unwraps it so `data.value` is `T | null` directly.
 *
 * @example
 * const { data } = await useKunFetch<HomeData>('/home')
 * // data.value?.galgames
 *
 * @example
 * const { data } = await useKunFetch<{ items: T[], total: number }>(
 *   '/topic',
 *   { query: pageData }
 * )
 */
export const useKunFetch = createUseFetch({
  // SSR-only ceiling so a stalled fetch can't hang the render and leak its
  // context. import.meta.server is build-time constant → client bundle gets
  // `undefined` (no timeout). See SSR_API_TIMEOUT_MS.
  timeout: import.meta.server ? SSR_API_TIMEOUT_MS : undefined,
  baseURL: computed(() => {
    const config = useRuntimeConfig()
    const base = import.meta.server
      ? config.apiBaseUrl
      : config.public.apiBaseUrl
    return `${base}/api`
  }),
  credentials: 'include',
  // SSR is cross-origin from Nuxt → Go API, so `credentials: 'include'` is
  // a no-op on the server. Manually forward the session cookie so per-user
  // flags (isLiked / isFavorited / isUpvoted) render correctly on first
  // paint. Filter to JUST the session cookie — see SESSION_COOKIE_NAME
  // comment above for why bundling everything triggers
  // "Request Header Fields Too Large".
  onRequest({ options }) {
    if (import.meta.server) {
      const session = extractSessionCookie(
        useRequestHeaders(['cookie']).cookie
      )
      if (session) {
        const merged = new Headers(options.headers as HeadersInit | undefined)
        merged.set('cookie', session)
        options.headers = merged
      }
    }
  },
  async onResponseError({ response }) {
    const resp = response._data as KunApiResponse<unknown> | undefined
    if (resp && resp.code !== 0) {
      await handleApiError(resp.code, resp.message)
    }
  },
  transform(resp: unknown) {
    const envelope = resp as KunApiResponse<unknown> | null | undefined
    if (!envelope || envelope.code !== 0) {
      return null
    }
    // Go's response.OKMessage omits `data`. Callers typically gate optimistic
    // updates on `if (result)`, so returning the message keeps the truthy
    // success signal intact while still distinguishing from the `null`
    // returned for real failures.
    return envelope.data !== undefined ? envelope.data : envelope.message
  }
})

/**
 * kunFetch — Imperative fetch for mutations (button clicks, form submits).
 * Client-side only. Unwraps { code, data } and handles errors.
 * Returns the unwrapped data, or null on error.
 *
 * @example
 * const result = await kunFetch<string>('/user/bio', {
 *   method: 'PUT',
 *   body: { bio: 'hello' }
 * })
 * if (result) { useMessage('更新成功', 'success') }
 */
export const kunFetch = async <T>(
  url: string,
  options?: Record<string, unknown>
): Promise<T | null> => {
  const config = useRuntimeConfig()
  // Prefer the server-only base URL when running on SSR — it bypasses
  // the public proxy and goes straight to the Go backend.
  const apiBase = import.meta.server
    ? `${config.apiBaseUrl}/api`
    : `${config.public.apiBaseUrl}/api`

  // Cookie forwarding (SSR): forward only the session cookie. Bundling
  // the whole cookie header (Pinia persisted stores, color-mode, etc.)
  // can blow past Fiber's 4KB header limit — same rationale as
  // useKunFetch above.
  const headers = new Headers(
    (options as { headers?: HeadersInit } | undefined)?.headers
  )
  if (import.meta.server) {
    const session = extractSessionCookie(
      useRequestHeaders(['cookie']).cookie
    )
    if (session) {
      headers.set('cookie', session)
    }
  }

  try {
    const resp = await $fetch<KunApiResponse<T>>(`${apiBase}${url}`, {
      // SSR-only default (see SSR_API_TIMEOUT_MS); listed before ...options so a
      // caller can still pass its own timeout (e.g. a long upload).
      timeout: import.meta.server ? SSR_API_TIMEOUT_MS : undefined,
      credentials: 'include',
      ...options,
      headers
    })

    if (!resp || resp.code !== 0) {
      if (resp) {
        await handleApiError(resp.code, resp.message)
      }
      return null
    }

    // Same fallback rationale as useKunFetch above: OKMessage responses
    // have no `data`, but callers check `if (result)` to confirm success.
    return resp.data !== undefined ? resp.data : (resp.message as T)
  } catch (error) {
    if (import.meta.client) {
      // ofetch rejects on ANY non-2xx status, so an app error that the Go
      // backend serves with an error status — e.g. an expired session is
      // 401 + { code: 205 } (errors.ErrAuthExpired) — lands here, NOT in the
      // `resp.code !== 0` branch above. Without unwrapping it we'd only ever
      // show "网络请求失败" and never reset the user / redirect to login.
      // ofetch attaches the parsed response body to FetchError.data, so route
      // it through the same handler as a 2xx-wrapped error; fall back to the
      // generic message only for genuine transport failures (no envelope).
      const resp = (error as { data?: KunApiResponse<unknown> }).data
      if (resp && resp.code !== 0) {
        await handleApiError(resp.code, resp.message)
      } else {
        useMessage('网络请求失败，请稍后重试', 'error')
      }
    }
    return null
  }
}
