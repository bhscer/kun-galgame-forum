export interface SectionTopic {
  id: number
  title: string
  content: string
  view: number
  likeCount: number
  replyCount: number
  hasBestAnswer: boolean
  isNSFWTopic: boolean
  user: KunUser
  created: Date | string
}

// Envelope for `GET /section`. Mirrors BE `SectionTopicsResponse`
// (apps/api/internal/section/dto). Container.vue previously referenced
// this name without an export, leaving the `useKunFetch` generic as `{}`.
export interface SectionTopicList {
  topics: SectionTopic[]
  total: number
}
