// Reaction keys + glyphs. Keys are in lock-step with the backend allowlist
// (reactionKeys in apps/api/.../topic_write_service.go). The glyphs are system
// emoji placeholders — Step 4 swaps them for the animated .webp assets.

export interface KunReactionMeta {
  key: string
  emoji: string
  label: string
}

// The two effectful reactions, pinned to the top of the picker.
export const KUN_REACTION_LIKE = 'like'
export const KUN_REACTION_DISLIKE = 'dislike'

export const KUN_REACTION_LIKE_NOTE = '点赞会使被点赞的用户萌萌点 +1'
export const KUN_REACTION_DISLIKE_NOTE = '被点踩多的用户将会被智乃屠杀'

// The plain emoji reactions shown in the picker grid (16).
export const KUN_REACTION_EMOJIS: KunReactionMeta[] = [
  { key: 'heart', emoji: '❤️', label: '爱心' },
  { key: 'fire', emoji: '🔥', label: '火' },
  { key: 'party', emoji: '🎉', label: '庆祝' },
  { key: 'love', emoji: '🥰', label: '喜欢' },
  { key: 'clap', emoji: '👏', label: '鼓掌' },
  { key: 'grin', emoji: '😁', label: '笑' },
  { key: 'thinking', emoji: '🤔', label: '思考' },
  { key: 'mindblown', emoji: '🤯', label: '震惊' },
  { key: 'scream', emoji: '😱', label: '尖叫' },
  { key: 'cry', emoji: '😢', label: '哭' },
  { key: 'pray', emoji: '🙏', label: '感谢' },
  { key: 'eyes', emoji: '👀', label: '关注' },
  { key: 'hundred', emoji: '💯', label: '满分' },
  { key: 'rofl', emoji: '🤣', label: '爆笑' },
  { key: 'partyface', emoji: '🥳', label: '派对' },
  { key: 'starstruck', emoji: '🤩', label: '星星眼' }
]

// Animated .webp asset for a reaction key (compressed Telegram emoji in
// public/emoji/). Used by the picker + bar; the glyph above is the alt/fallback.
export const reactionAsset = (key: string) => `/emoji/${key}.webp`

// key → glyph, for rendering a reaction chip in the bar (includes like/dislike).
export const KUN_REACTION_EMOJI: Record<string, string> = {
  like: '👍',
  dislike: '👎',
  heart: '❤️',
  fire: '🔥',
  party: '🎉',
  love: '🥰',
  clap: '👏',
  grin: '😁',
  thinking: '🤔',
  mindblown: '🤯',
  scream: '😱',
  cry: '😢',
  pray: '🙏',
  eyes: '👀',
  hundred: '💯',
  rofl: '🤣',
  partyface: '🥳',
  hearteyes: '😍'
}
