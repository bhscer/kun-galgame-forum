// Shared "kick off an OAuth flow" entry points. Both login and register
// generate the same PKCE bundle + write the same sessionStorage keys
// (oauth_code_verifier / oauth_state) that pages/auth/callback.vue
// reads back when the browser returns from OAuth. The only difference
// is WHERE the browser lands first:
//
//   - login    → ${oauthServerUrl}/oauth/authorize?<params>
//   - register → ${oauthFrontendUrl}/auth/register?redirect=<encoded(authorize)>
//                (per docs/oauth/05-registration.md — OAuth's register
//                 page chains back to /oauth/authorize after the form,
//                 silently consent-skipped for first-party clients)
//
// Use these from the top-bar AuthModal, from auth-expired toast handlers,
// or from anywhere else that needs to start the OAuth dance. They
// hard-redirect the page (window.location.href), so the caller doesn't
// need to await anything beyond the URL build.

const buildAuthorizeUrl = async (): Promise<string> => {
  const config = useRuntimeConfig()

  const codeVerifier = generateCodeVerifier()
  const codeChallenge = await generateCodeChallenge(codeVerifier)
  const state = generateState()

  // Persist the per-attempt secrets so /auth/callback can validate the
  // returned `state` and complete the PKCE exchange. Both entry points
  // share the same keys so the callback never has to know which one
  // started the flow.
  sessionStorage.setItem('oauth_code_verifier', codeVerifier)
  sessionStorage.setItem('oauth_state', state)

  const params = new URLSearchParams({
    client_id: config.public.oauthClientId as string,
    redirect_uri: config.public.oauthRedirectUri as string,
    response_type: 'code',
    scope: 'openid profile',
    state,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256'
  })
  return `${config.public.oauthServerUrl}/oauth/authorize?${params}`
}

export const startOAuthLogin = async (): Promise<void> => {
  const authorizeUrl = await buildAuthorizeUrl()
  window.location.href = authorizeUrl
}

export const startOAuthRegister = async (): Promise<void> => {
  const config = useRuntimeConfig()
  const authorizeUrl = await buildAuthorizeUrl()
  const registerUrl = `${config.public.oauthFrontendUrl}/auth/register?redirect=${encodeURIComponent(authorizeUrl)}`
  window.location.href = registerUrl
}
