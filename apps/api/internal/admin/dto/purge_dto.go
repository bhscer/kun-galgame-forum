package dto

// UserContentStats is the per-type breakdown of a user's kungal content,
// used both to preview a purge (GET) and to report what was deleted (DELETE).
type UserContentStats struct {
	Topics           int64 `json:"topics"`
	Replies          int64 `json:"replies"`
	TopicComments    int64 `json:"topicComments"`
	GalgameComments  int64 `json:"galgameComments"`
	Ratings          int64 `json:"ratings"`
	RatingComments   int64 `json:"ratingComments"`
	Resources        int64 `json:"resources"`
	Websites         int64 `json:"websites"`
	WebsiteComments  int64 `json:"websiteComments"`
	Toolsets         int64 `json:"toolsets"`
	ToolsetResources int64 `json:"toolsetResources"`
	ToolsetComments  int64 `json:"toolsetComments"`
	ChatMessages     int64 `json:"chatMessages"`
	Messages         int64 `json:"messages"`
	Interactions     int64 `json:"interactions"`
	Total            int64 `json:"total"`
}
