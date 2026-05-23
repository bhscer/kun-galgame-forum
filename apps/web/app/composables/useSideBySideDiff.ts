// Side-by-side scalar diff. Given an old + new string, returns two
// HTML fragments:
//
//   - `oldHtml`: the full OLD string, with characters that were
//     REMOVED in the new version wrapped in <strong>.
//   - `newHtml`: the full NEW string, with characters that were
//     ADDED in the new version wrapped in <b>.
//
// Pair this with red styling on <strong> and green styling on <b>
// (matches the existing kun-diff convention) and you get a
// GitHub-style side-by-side view: the left column is the original
// text with strikethrough on what disappeared, the right column is
// the new text with highlight on what got added.
//
// Algorithm: classic LCS DP — same as the existing useDiff. The only
// difference is that we keep `inLcsA` / `inLcsB` membership flags per
// character instead of weaving a single interleaved string, so each
// side can render the full original/new text independently.
//
// HTML escaping is mandatory: this output is consumed via `v-html`
// after DOMPurify, but escaping at the source means user input that
// happens to contain `<` / `&` / `"` won't accidentally end up inside
// our own `<b>` / `<strong>` wrappers as broken markup that DOMPurify
// then has to clean up.

const escapeHtml = (ch: string): string => {
  switch (ch) {
    case '&':
      return '&amp;'
    case '<':
      return '&lt;'
    case '>':
      return '&gt;'
    case '"':
      return '&quot;'
    case "'":
      return '&#39;'
    default:
      return ch
  }
}

export const useSideBySideDiff = (
  oldStr: string,
  newStr: string
): { oldHtml: string; newHtml: string } => {
  const a = oldStr ?? ''
  const b = newStr ?? ''

  // Standard LCS DP table.
  const dp: number[][] = []
  for (let i = 0; i <= a.length; i++) {
    dp[i] = []
    for (let j = 0; j <= b.length; j++) {
      if (i === 0 || j === 0) dp[i]![j] = 0
      else if (a[i - 1] === b[j - 1]) dp[i]![j] = dp[i - 1]![j - 1]! + 1
      else dp[i]![j] = Math.max(dp[i - 1]![j]!, dp[i]![j - 1]!)
    }
  }

  // Walk the table to mark which characters of A and B are part of
  // the longest common subsequence. Everything NOT marked is a true
  // diff char on its side.
  const inLcsA = new Array<boolean>(a.length).fill(false)
  const inLcsB = new Array<boolean>(b.length).fill(false)
  let i = a.length
  let j = b.length
  while (i > 0 && j > 0) {
    if (a[i - 1] === b[j - 1]) {
      inLcsA[i - 1] = true
      inLcsB[j - 1] = true
      i--
      j--
    } else if (dp[i - 1]![j]! > dp[i]![j - 1]!) {
      i--
    } else {
      j--
    }
  }

  let oldHtml = ''
  for (let k = 0; k < a.length; k++) {
    const ch = escapeHtml(a[k]!)
    oldHtml += inLcsA[k] ? ch : `<strong>${ch}</strong>`
  }

  let newHtml = ''
  for (let k = 0; k < b.length; k++) {
    const ch = escapeHtml(b[k]!)
    newHtml += inLcsB[k] ? ch : `<b>${ch}</b>`
  }

  return { oldHtml, newHtml }
}
