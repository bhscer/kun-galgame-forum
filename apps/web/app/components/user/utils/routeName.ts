export interface Nav {
  name: string
  router: string
  redirect?: string
  collapsed?: boolean
  permission?: number[]
  child?: Nav[]
}

export const navBarRoute: Ref<Nav[]> = ref([
  {
    name: 'profile',
    router: 'info',
    permission: [1, 2, 3, 4]
  },
  {
    name: 'setting',
    router: 'setting',
    permission: [4]
  },
  // Email + password edits moved to OAuth profile per the 2026-05-23
  // policy (docs/oauth/README.md + 02-user-profile.md §身份层操作 vs
  // 展示操作). The setting page surfaces redirect buttons — no kungal-
  // side route for either.
  {
    name: 'topic',
    permission: [1, 2, 3, 4],
    redirect: 'topic/topic',
    router: 'topic'
  },
  {
    name: 'galgame',
    permission: [1, 2, 3, 4],
    redirect: 'galgame/galgame-like',
    router: 'galgame'
  },
  {
    name: 'rating',
    router: 'rating',
    permission: [1, 2, 3, 4]
  },
  {
    name: 'resource',
    permission: [1, 2, 3, 4],
    redirect: 'resource/valid',
    router: 'resource'
  },
  {
    name: 'reply',
    router: 'reply',
    redirect: 'reply/reply-created',
    permission: [1, 2, 3, 4]
  },
  {
    name: 'comment',
    router: 'comment',
    redirect: 'comment/comment-created',
    permission: [1, 2, 3, 4]
  }
])
