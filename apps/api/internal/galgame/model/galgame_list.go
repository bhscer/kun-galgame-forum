package model

// GalgameListFilter is the parameter bundle for the galgame list repository.
type GalgameListFilter struct {
	Type                 string
	Language             string
	Platform             string
	SortField            string
	SortOrder            string
	IncludeProviders     []string
	ExcludeOnlyProviders []string
	// Release-date range, already resolved to inclusive "YYYY-MM-DD"
	// boundaries by utils.ParseReleaseLowerBound/UpperBound (empty =
	// no bound on that side). Compared against galgame.release_date;
	// NULL rows drop out once either bound is set (wiki §17.4).
	ReleasedFrom string
	ReleasedTo   string
	Page         int
	Limit        int
}

// GalgameResourceMeta holds a platform/language tuple from galgame_resource,
// used when aggregating per-galgame platform/language sets.
type GalgameResourceMeta struct {
	GalgameID int    `gorm:"column:galgame_id"`
	Platform  string `gorm:"column:platform"`
	Language  string `gorm:"column:language"`
}
