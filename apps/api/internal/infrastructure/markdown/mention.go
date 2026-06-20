package markdown

import (
	"regexp"
	"strconv"
)

// mentionIDRe extracts the user id from a stored mention token
// [@name](kungal-user:<id>). It runs against the raw markdown SOURCE (pre-render),
// so it sees the token before the render transform rewrites it to a link.
var mentionIDRe = regexp.MustCompile(`kungal-user:(\d+)`)

// ExtractMentionIDs returns the unique, positive user ids @mentioned in raw
// markdown `content`, in first-seen order. Used to fan out mention notifications
// and to batch-resolve current display names at render time.
func ExtractMentionIDs(content string) []int {
	matches := mentionIDRe.FindAllStringSubmatch(content, -1)
	if matches == nil {
		return nil
	}
	seen := make(map[int]bool, len(matches))
	ids := make([]int, 0, len(matches))
	for _, m := range matches {
		id, err := strconv.Atoi(m[1])
		if err != nil || id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}
