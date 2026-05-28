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
	// Discontinuous month set (1–12), AND-combined with the year range
	// (wiki §17.10): keep only games whose release month ∈ this set,
	// across all in-range years. Empty = no month filter.
	ReleasedMonths []int
	// Bayesian-rating advanced filters (Design A — computed live from a
	// galgame_rating aggregation join; no denormalized column). The
	// smoothing constants (prior C, global mean m) live inside the repo.
	//   MinRatingCount — keep galgames with at least this many ratings
	//   MinRating      — keep galgames whose Bayesian score >= this (0–10)
	// Zero = filter inactive. Rating SORT is driven by SortField=="rating".
	MinRatingCount int
	MinRating      float64
	Page           int
	Limit          int
}

// GalgameResourceMeta holds a platform/language tuple from galgame_resource,
// used when aggregating per-galgame platform/language sets.
type GalgameResourceMeta struct {
	GalgameID int    `gorm:"column:galgame_id"`
	Platform  string `gorm:"column:platform"`
	Language  string `gorm:"column:language"`
}
