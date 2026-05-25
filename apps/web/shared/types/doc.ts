export interface DocTocLink {
  id: string
  text: string
  depth: number
  children?: DocTocLink[]
}

export interface DocCategoryItem {
  id: number
  slug: string
  title: string
  description: string
  icon: string
  sort_order: number
  created: Date | string
  updated: Date | string
}

export interface DocCategoryListResponse {
  categories: DocCategoryItem[]
  total: number
  page: number
  limit: number
}

export interface DocTagItem {
  id: number
  slug: string
  title: string
  description: string
  created: Date | string
  updated: Date | string
}

export interface DocTagListResponse {
  tags: DocTagItem[]
  total: number
  page: number
  limit: number
}

export interface DocArticleCategoryBrief {
  id: number
  slug: string
  title: string
}

export interface DocArticle {
  id: number
  title: string
  slug: string
  path: string
  description: string
  banner: string
  status: number
  isPin: boolean
  view: number
  publishedTime: Date | string
  editedTime: Date | string | null
  contentMarkdown: string
  categoryId: number
  authorId: number
  category: DocArticleCategoryBrief
  created: Date | string
  updated: Date | string
  contentHtml?: string
  toc?: DocTocLink[]
}

// Kept for backward compatibility with components that use the old shape
export type DocArticleSummary = DocArticle
export type DocArticleDetail = DocArticle

export interface DocArticleListResponse {
  items: DocArticle[]
  total: number
}
