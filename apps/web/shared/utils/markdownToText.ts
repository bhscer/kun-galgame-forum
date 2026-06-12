// Strip Markdown syntax down to readable plain text. Two modes:
//   • default — flatten to a SINGLE line (newlines → spaces). Right for SEO
//     meta descriptions and one-line list previews, which is almost every
//     caller.
//   • { preserveNewlines: true } — keep the author's line breaks (only blank-
//     line runs are collapsed). Used by the resource-note card, whose <p>
//     renders with `whitespace-pre-line`; without this its line breaks were
//     flattened away ("备注换行失效").
export const markdownToText = (
  markdown: string,
  opts?: { preserveNewlines?: boolean }
) => {
  if (!markdown) return ''
  const stripped = markdown
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replace(/(\*\*|__)(.*?)\1/g, '$2')
    .replace(/(\*|_)(.*?)\1/g, '$2')
    .replace(/^\s*(#{1,6})\s+(.*)/gm, '$2')
    .replace(/`/g, '')
    .replace(/^(-{3,}|\*{3,})$/gm, '')
    .replace(/^\s*([-*+]|\d+\.)\s+/gm, '')
  return (
    opts?.preserveNewlines
      ? stripped.replace(/[ \t]+$/gm, '').replace(/\n{3,}/g, '\n\n')
      : stripped.replace(/\n+/g, ' ')
  ).trim()
}
