export default defineNuxtRouteMiddleware(() => {
  const { id } = usePersistUserStore()

  // Bounce unauthenticated traffic to the homepage and prompt the user
  // to click the top-bar 登录 (which opens KunAuthModal → OAuth jump).
  // We don't auto-kick the OAuth flow here because middleware runs in
  // both SSR and client contexts, and a forced window.location away
  // would surprise users mid-navigation. The toast tells them what
  // happened; the next click is theirs to make.
  if (!id) {
    useMessage(10249, 'warn', 5000)
    return navigateTo('/')
  }
})
