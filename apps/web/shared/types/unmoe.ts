export interface UnmoeLog {
  id: number
  user: KunUser
  name: string
  description: KunLanguage
  // BE `Result` is plain string (apps/api/internal/unmoe/dto). The old
  // `string | number` union let stale FE callers pass numbers in;
  // narrowing here so the next consumer doesn't have to widen handling.
  result: string
  created: Date | string
}
