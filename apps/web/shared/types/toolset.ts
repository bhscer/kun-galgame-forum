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
  // BE returns `null` when the caller hasn't rated yet (PracticalityResponse.Mine *int).
  mine: number | null
}

// Server-driven init response from the artifact service (via kungal's BFF).
// When multipart is false the browser does one PUT to uploadUrl; otherwise it
// slices by partSize and PUTs each part to parts[i].url, collecting ETags.
export interface ToolsetUploadInitResponse {
  artifactUuid: string
  multipart: boolean
  uploadUrl?: string
  partSize?: number
  parts?: {
    partNumber: number
    url: string
  }[]
  expiresAt: string
}

export interface ToolsetUploadCompleteResponse {
  artifactUuid: string
  size: number
}

// Result emitted from the S3 upload widget once a full upload (init → PUT →
// complete) succeeds. The artifact uuid binds the upload to a toolset_resource
// row at create time; size pre-fills the file size input.
export interface ToolsetUploadResult {
  artifactUuid: string
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
