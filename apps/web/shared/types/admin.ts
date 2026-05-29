export interface AdminOverStats {
  date: string
  [key: string]: number | string
}

// Per-type breakdown of a user's kungal content (BE: admin/dto
// UserContentStats). Used to preview and report a content purge.
export interface AdminUserContentStats {
  topics: number
  replies: number
  topicComments: number
  galgameComments: number
  ratings: number
  ratingComments: number
  resources: number
  websites: number
  websiteComments: number
  toolsets: number
  toolsetResources: number
  toolsetComments: number
  interactions: number
  total: number
}
