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

import Cookies from 'js-cookie'

const buildAuthorizeUrl = async (): Promise<string> => {
  const config = useRuntimeConfig()

  const codeVerifier = generateCodeVerifier()
  const codeChallenge = await generateCodeChallenge(codeVerifier)
  const state = generateState()

  // Persist the per-attempt secrets so /auth/callback can validate the
  // returned `state` and complete the PKCE exchange.
  //
  // Stored in COOKIES, not sessionStorage: sessionStorage is per-tab and is NOT
  // reliably preserved across the cross-origin OAuth redirect on lightweight /
  // older mobile browsers (Via, in-app WebViews, COOP / process switches, or
  // when the auth page opens in a new tab). The callback then reads null and
  // fails with "State 不匹配". Cookies are origin-scoped, survive the round-trip,
  // and auto-expire. 10 min = the auth round-trip window; `secure` only on https
  // so dev over http still sets them. Both entry points (login / register) write
  // the same keys so the callback never has to know which one started the flow.
  const oauthCookieOptions = {
    expires: 10 / 1440,
    sameSite: 'lax' as const,
    secure: location.protocol === 'https:',
    path: '/'
  }
  Cookies.set('oauth_code_verifier', codeVerifier, oauthCookieOptions)
  Cookies.set('oauth_state', state, oauthCookieOptions)

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
