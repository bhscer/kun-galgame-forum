// uploadGalgameImage — POST /image/galgame proxy helper.
//
// K-PR3a infrastructure: covers/screenshots reference image_service by
// hash. Before the user can add a new cover/screenshot row to a galgame
// PUT/PR payload, they must upload the file via this endpoint to obtain
// the hash. K-PR3b's cover/screenshot editors call this on file-pick.
//
// The endpoint expects multipart/form-data:
//   - file:   the image binary
//   - preset: one of "galgame_banner" / "galgame_screenshot"
//             (kungal handler enforces the allowlist; passing anything
//              else → 400 with a clear message)
//
// The success payload mirrors image_service's /image/upload response,
// adapted to camelCase by kungal (see UploadGalgameResult in
// apps/api/internal/image/service/galgame_upload.go). `url` is the main
// CDN URL — the editor can render it immediately and stash `hash` in
// the cover/screenshot row.
//
// kunFetch surfaces business errors via its response handler (quota
// exceeded, moderation rejected, missing credentials, etc.); on those
// it returns null. Callers should treat null as "show wiki message,
// don't proceed" and not retry blindly.

export type GalgameImageUploadPreset = 'galgame_banner' | 'galgame_screenshot'

export interface UploadGalgameImageResult {
  hash: string
  url: string
  width: number
  height: number
  sizeBytes: number
  variantUrls?: Record<string, string>
  deduplicated: boolean
}

export const uploadGalgameImage = async (
  file: Blob,
  preset: GalgameImageUploadPreset,
  filename = 'image'
): Promise<UploadGalgameImageResult | null> => {
  const form = new FormData()
  form.append('file', file, filename)
  form.append('preset', preset)
  return await kunFetch<UploadGalgameImageResult>('/image/galgame', {
    method: 'POST',
    body: form
  })
}
