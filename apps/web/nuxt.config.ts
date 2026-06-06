import tailwindcss from '@tailwindcss/vite'
import fs from 'fs'
import path from 'path'
import { ICON_COLLECTIONS, ICON_NAMES } from './lib/icon'
import type { TSConfig } from 'pkg-types'

const packageJson = JSON.parse(
  fs.readFileSync(path.resolve(__dirname, '..', '..', 'package.json'), 'utf-8')
)
const appVersion = packageJson.version

console.log(appVersion)

const sharedTsConfig: TSConfig = {
  exclude: ['**/backup/**', '**/dist/**', '**/node_modules/**']
}

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // @kun/ui is consumed as a Nuxt layer — components / composables /
  // styles flow in via the layer system instead of duplicated files.
  extends: ['../../packages/ui'],

  // Alias for explicit type-only imports from the layer (the auto-import
  // path handles components / composables, but `import type {…}` still
  // needs a real path).
  alias: {
    '@kun/ui': path.resolve(__dirname, '../../packages/ui/app')
  },

  devtools: { enabled: false },

  app: {
    pageTransition: { name: 'kun-page', mode: 'out-in' },
    layoutTransition: { name: 'kun-page', mode: 'out-in' },

    // https://github.com/nuxt/nuxt/issues/26565#issuecomment-3448517709
    baseURL: '/'
    // buildAssetsDir: `/_nuxt/v${Math.floor(Date.now() / 1000).toString()}/`
  },

  experimental: {
    scanPageMeta: true,
    typescriptPlugin: true
  },

  compatibilityDate: '2025-07-15',

  devServer: {
    host: '127.0.0.1',
    port: 2333
  },

  modules: [
    '@nuxt/image',
    '@nuxt/icon',
    '@nuxt/eslint',
    '@nuxtjs/color-mode',
    '@pinia/nuxt',
    '@dxup/nuxt',
    'pinia-plugin-persistedstate/nuxt',
    'nuxt-schema-org',
    'nuxt-umami'
  ],

  runtimeConfig: {
    // Server-only: used by useFetch during SSR to directly reach Go API
    apiBaseUrl: process.env.API_BASE_URL || 'http://127.0.0.1:2334',

    public: {
      KUN_GALGAME_URL: process.env.KUN_GALGAME_URL,
      KUN_VISUAL_NOVEL_FORUM_YANDEX_VERIFICATION:
        process.env.KUN_VISUAL_NOVEL_FORUM_YANDEX_VERIFICATION,
      KUN_VISUAL_NOVEL_VERSION: appVersion,

      // Go API base URL (client-side, relative — goes through browser)
      apiBaseUrl: process.env.API_BASE_URL || 'http://127.0.0.1:2334',

      // OAuth
      // oauthServerUrl   : API base (e.g. .../api/v1) — used for HTTP calls.
      // oauthFrontendUrl : web-app root — used for cross-app navigation
      //                    (改邮箱 / 改密码 jumps to /profile). Kept
      //                    distinct because dev runs them on separate
      //                    ports; in prod they share a domain.
      oauthServerUrl:
        process.env.OAUTH_SERVER_URL || 'http://127.0.0.1:9277/api/v1',
      oauthFrontendUrl:
        process.env.OAUTH_FRONTEND_URL || 'https://oauth.kungal.com',
      oauthClientId: process.env.OAUTH_CLIENT_ID || '',
      oauthRedirectUri:
        process.env.OAUTH_REDIRECT_URI || 'http://127.0.0.1:2333/auth/callback',

      // Galgame Wiki Service
      galgameWikiUrl:
        process.env.GALGAME_WIKI_URL || 'http://127.0.0.1:9280/api'
    }
  },

  routeRules: {},

  imports: {
    dirs: ['./composables', './config', './utils']
  },

  site: {
    url: process.env.KUN_GALGAME_URL
  },

  umami: {
    id: process.env.KUN_VISUAL_NOVEL_FORUM_UMAMI_ID,
    host: 'https://umami.kungal.org/',
    autoTrack: true
  },

  // Frontend
  css: ['~/styles/index.css'],

  icon: {
    mode: 'svg',
    // The default local endpoint is /api/_nuxt_icon, but in prod Traefik routes
    // ALL of /api/* to the Go API — so @nuxt/icon's client fallback for any icon
    // not in clientBundle (the ~20 dynamic `:name=expr` bindings) hit Go and
    // 404'd, breaking those icons and spamming the API error log. Move it off
    // the Go-owned /api namespace so it reaches Nitro, which serves it from
    // serverBundle (which carries every collection, so dynamic icons resolve).
    localApiEndpoint: '/_nuxt_icon',
    serverBundle: {
      collections: ICON_COLLECTIONS
    },
    clientBundle: {
      icons: ICON_NAMES,
      scan: false
    }
  },

  typescript: {
    tsConfig: {
      ...sharedTsConfig
    },
    nodeTsConfig: {
      ...sharedTsConfig
    },
    sharedTsConfig: {
      ...sharedTsConfig
    }
  },

  vite: {
    plugins: [tailwindcss()]
    // optimizeDeps: {
    //   include: ['isomorphic-dompurify', 'date-fns/locale']
    // }
  },

  // $production: {
  //   vite: {
  //     esbuild: {
  //       drop: ['console', 'debugger']
  //     }
  //   }
  // },

  // ogImage: {
  //   fonts: [
  //     {
  //       name: 'Lolita',
  //       weight: 400,
  //       path: '/fonts/Lolita.ttf'
  //     }
  //   ]
  // },

  eslint: {
    config: {
      stylistic: false
    }
  },

  // Pinia store functions auto imports
  pinia: {
    storesDirs: ['./store/**']
  },

  piniaPluginPersistedstate: {
    cookieOptions: {
      maxAge: 60 * 60 * 24 * 7,
      sameSite: 'strict'
    }
  },

  colorMode: {
    preference: 'system',
    fallback: 'light',
    globalName: '__KUNGALGAME_COLOR_MODE__',
    componentName: 'ColorScheme',
    classPrefix: 'kun-',
    classSuffix: '-mode',
    storageKey: 'kungalgame-color-mode'
  },

  nitro: {
    // Nitro's default esbuild target (es2019) rejects top-level await, which
    // utils/sanitize.ts uses to load jsdom server-only. The prod runtime is
    // node 24, so bump the server target to es2022 (which supports TLA).
    esbuild: {
      options: {
        target: 'es2022'
      }
    },
    typescript: {
      tsConfig: {
        ...sharedTsConfig
      }
    }
  }
})
