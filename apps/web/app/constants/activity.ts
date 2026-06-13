export const KUN_ACTIVITY_TYPE_TYPE: Record<string, string> = {
  GALGAME_CREATION: 'Galgame',
  GALGAME_RATING_CREATION: 'Galgame 评分',
  GALGAME_RATING_COMMENT_CREATION: 'Galgame 评分评论',
  TOPIC_CREATION: '新话题',
  MESSAGE_UPVOTE: '话题被推',
  MESSAGE_SOLUTION: '最佳答案',
  TOPIC_REPLY_CREATION: '话题回复',
  TOPIC_COMMENT_CREATION: '话题评论',
  GALGAME_WEBSITE_CREATION: 'Galgame 网站',
  GALGAME_RESOURCE_CREATION: 'Galgame 资源',
  GALGAME_EDIT: 'Galgame 编辑',
  GALGAME_PR_CREATION: '提出更新请求',
  GALGAME_COMMENT_CREATION: 'Galgame 评论',
  GALGAME_WEBSITE_COMMENT_CREATION: 'Galgame 网站评论',
  TODO_CREATION: '待办',
  UPDATE_LOG_CREATION: '更新日志',
  TOOLSET_CREATION: 'Galgame 工具',
  TOOLSET_RESOURCE_CREATION: '工具资源',
  TOOLSET_COMMENT_CREATION: '工具评论'
}

export const KUN_ALLOWED_ACTIVITY_TYPE = [
  'GALGAME_CREATION',
  'GALGAME_COMMENT_CREATION',
  'GALGAME_RATING_CREATION',
  'GALGAME_RATING_COMMENT_CREATION',
  'GALGAME_WEBSITE_CREATION',
  'GALGAME_WEBSITE_COMMENT_CREATION',
  'GALGAME_RESOURCE_CREATION',
  'GALGAME_EDIT',
  'GALGAME_PR_CREATION',
  'TOOLSET_CREATION',
  'TOOLSET_RESOURCE_CREATION',
  'TOOLSET_COMMENT_CREATION',
  'TOPIC_CREATION',
  'TOPIC_REPLY_CREATION',
  'TOPIC_COMMENT_CREATION',
  'TODO_CREATION',
  'UPDATE_LOG_CREATION',
  'MESSAGE_UPVOTE',
  'MESSAGE_SOLUTION'
] as const

// Grouped view of the activity types for the /activity/category picker. A flat
// 18-item tab row was unscannable and overflowed, so the picker is a grouped
// dropdown instead. INVARIANT: every key in KUN_ACTIVITY_TYPE_TYPE must appear
// in exactly one group's `types` (ordering within a group is the display order).
export const KUN_ACTIVITY_GROUPS: { label: string; types: string[] }[] = [
  {
    label: 'Galgame',
    types: [
      'GALGAME_CREATION',
      'GALGAME_EDIT',
      'GALGAME_PR_CREATION',
      'GALGAME_RESOURCE_CREATION',
      'GALGAME_RATING_CREATION',
      'GALGAME_RATING_COMMENT_CREATION',
      'GALGAME_COMMENT_CREATION',
      'GALGAME_WEBSITE_CREATION',
      'GALGAME_WEBSITE_COMMENT_CREATION'
    ]
  },
  {
    label: '社区',
    types: [
      'TOPIC_CREATION',
      'TOPIC_REPLY_CREATION',
      'TOPIC_COMMENT_CREATION',
      'MESSAGE_UPVOTE',
      'MESSAGE_SOLUTION'
    ]
  },
  {
    label: '工具集',
    types: [
      'TOOLSET_CREATION',
      'TOOLSET_RESOURCE_CREATION',
      'TOOLSET_COMMENT_CREATION'
    ]
  },
  {
    label: '站务',
    types: ['TODO_CREATION', 'UPDATE_LOG_CREATION']
  }
]

export const KUN_ACTIVITY_ICON_MAP: Record<string, string> = {
  GALGAME_CREATION: 'lucide:gamepad-2',
  GALGAME_RATING_CREATION: 'lucide:star',
  GALGAME_RATING_COMMENT_CREATION: 'lucide:message-square-text',
  GALGAME_COMMENT_CREATION: 'lucide:message-square',
  GALGAME_WEBSITE_CREATION: 'lucide:globe',
  GALGAME_WEBSITE_COMMENT_CREATION: 'lucide:message-square-text',
  GALGAME_RESOURCE_CREATION: 'lucide:box',
  GALGAME_EDIT: 'lucide:file-pen-line',
  GALGAME_PR_CREATION: 'lucide:git-pull-request',
  TOOLSET_CREATION: 'lucide:wrench',
  TOOLSET_RESOURCE_CREATION: 'lucide:package-plus',
  TOOLSET_COMMENT_CREATION: 'lucide:wrench',
  TOPIC_CREATION: 'icon-park-outline:topic',
  TOPIC_REPLY_CREATION: 'carbon:reply',
  TOPIC_COMMENT_CREATION: 'lucide:message-circle-more',
  TODO_CREATION: 'lucide:list-checks',
  UPDATE_LOG_CREATION: 'lucide:file-clock',
  MESSAGE_UPVOTE: 'lucide:sparkles',
  MESSAGE_SOLUTION: 'lucide:bookmark-check'
}
