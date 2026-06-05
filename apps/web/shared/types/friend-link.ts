export type FriendLinkCategory = 'official' | 'galgame' | 'others'

export type FriendLinkStatus = 'normal' | 'essential' | 'down'

export interface FriendLink {
  id: number
  category: FriendLinkCategory
  name: string
  link: string
  description: string
  /** Full image URL (image_service webp, or the legacy /friends/<name>.webp). */
  banner: string
  status: FriendLinkStatus
  sortOrder: number
  created: string
  updated: string
}

/** Response shape of GET /friend-link — links grouped by the 3 fixed categories. */
export interface GroupedFriendLinks {
  official: FriendLink[]
  galgame: FriendLink[]
  others: FriendLink[]
}

/** Create/update payload from the admin form (id present only when editing). */
export interface FriendLinkInput {
  id?: number
  category: FriendLinkCategory
  name: string
  link: string
  description: string
  banner: string
  status: FriendLinkStatus
}
