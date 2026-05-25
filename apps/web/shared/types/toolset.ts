import type { ToolsetComment } from './toolset-comment'

export interface ToolsetCard {
  id: number
  name: string
  user: KunUser
  type: string
  platform: string
  language: string
  version: string
  view: number
  download: number
  commentCount: number
  practicalityAvg: number | null
  resource_update_time: Date | string
}

export interface ToolsetDetail {
  id: number
  name: string
  contentHtml: string
  contentMarkdown: string
  type: string
  platform: string
  language: string
  version: string
  homepage: string[]
  download: number
  view: number
  user: KunUser
  aliases: string[]
  practicalityAvg: number | null
  practicalityCount: number
  resource_update_time: Date | string
  resource: ToolsetResource[]
  edited: Date | string | null
  created: Date | string
  updated: Date | string
  ratingCounts: Record<number, number>
  commentCount: number
  commentPreview: ToolsetComment[]
  contributors: KunUser[]
}

export interface ToolsetRating {
  counts: {
    [x: number]: number
  }
  avg: number
  mine: number
}

export interface ToolsetLargeFileUploadResponse {
  key: string
  salt: string
  uploadId: string
  parts: {
    partNumber: number
    presignedUrl: string
  }[]
}

export interface ToolsetSmallFileUploadResponse {
  key: string
  salt: string
  presignedUrl: string
}

export interface ToolsetUploadCompleteResponse {
  key: string
  size: number
}

// Result emitted from the S3 upload widget once a full upload (init → PUT
// → complete) succeeds. Combines the salt from the init step (the API only
// returns it there) with the key+size from the complete step. Downstream
// resource creation needs salt to bind the upload to a toolset_resource
// row, and size to pre-fill the file size input.
export interface ToolsetUploadResult {
  salt: string
  key: string
  size: number
}

export interface ToolsetResource {
  id: number
  type: string
  size: string
  download: number
  status: number
}

export interface ToolsetResourceDetail extends ToolsetResource {
  user: KunUser
  content: string
  code: string
  note: string
  password: string
  edited: Date | string | null
  created: Date | string
  updated: Date | string
}
