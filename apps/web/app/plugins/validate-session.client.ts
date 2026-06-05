// Reconcile persisted login state with the REAL session on each app load.
//
// usePersistUserStore is persisted independently of the `kungal_session` cookie,
// and BOTH the auth middleware and the 205→logout path key off that stale
// store.id — so a user whose session expired/invalidated (notably migrated
// users carrying a pre-migration token) keeps appearing logged-in until they
// happen to trigger an authed request. This pings the authed `/user/status`
// once on load: a dead session returns code 205, which kunFetch's
// handleApiError turns into resetUser() + redirect (auto-logout). A live
// session is a cheap no-op.
export default defineNuxtPlugin((nuxtApp) => {
  const userStore = usePersistUserStore()
  if (!userStore.id) return

  // Fire-and-forget so we don't delay hydration. runWithContext keeps the Nuxt
  // composables used inside kunFetch / handleApiError valid across the async
  // continuation (useRuntimeConfig / navigateTo / useMessage).
  nuxtApp.runWithContext(() => kunFetch('/user/status'))
})
