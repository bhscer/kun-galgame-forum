package repository

import (
	"strconv"
	"strings"

	"kun-galgame-api/internal/galgame/model"

	"gorm.io/gorm"
)

// bayesianPriorC is the confidence prior ("virtual votes") for the
// galgame rating Bayesian average: score = (C·m + Σoverall) / (C + n).
// Higher C pulls low-vote-count games harder toward the global mean.
// 10 is a sane default for kungal's modest rating volume — tune here.
const bayesianPriorC = 10.0

// ratingAggJoin pre-aggregates galgame_rating to one row per galgame
// (Σoverall, n). LEFT JOIN so unrated galgames still appear (rt.rsum /
// rt.rcnt NULL). Alias `rt` to avoid clashing with the `gr` galgame_
// resource alias used in the resource-filter branch.
const ratingAggJoin = "LEFT JOIN (SELECT galgame_id, SUM(overall) AS rsum, " +
	"COUNT(*) AS rcnt FROM galgame_rating GROUP BY galgame_id) rt " +
	"ON rt.galgame_id = g.id"

// GalgameListRepository owns the paginated list query with resource-filter
// support (type/language/platform + include/exclude provider sets).
type GalgameListRepository struct {
	db *gorm.DB
}

func NewGalgameListRepository(db *gorm.DB) *GalgameListRepository {
	return &GalgameListRepository{db: db}
}

func (r *GalgameListRepository) DB() *gorm.DB { return r.db }

// allProviders is the closed set of provider names used by the exclude-only filter.
var allProviders = []string{
	"baidu", "aliyun", "quark", "pan123", "tianyiyun",
	"caiyun", "xunlei", "uc", "lanzou", "other",
}

// ListIDs returns galgame IDs matching the given filter, paginated and sorted.
// If hasResourceFilter (returned by HasResourceFilter) is true, a JOIN against
// galgame_resource is used; otherwise a simple galgame-only scan.
func (r *GalgameListRepository) ListIDs(f model.GalgameListFilter) (ids []int, total int64) {
	sortCol := "g.updated"
	switch f.SortField {
	case "time":
		sortCol = "g.updated"
	case "created":
		sortCol = "g.created"
	case "view":
		sortCol = "g.view"
	case "release_date":
		sortCol = "g.release_date"
	}

	ratingSort := f.SortField == "rating"
	ratingFilter := f.MinRatingCount > 0 || f.MinRating > 0

	// Bayesian score expression (computed once: pulls the live global mean
	// m). Built only when a rating sort/filter is active so the AVG query
	// isn't paid otherwise.
	var bayes string
	if ratingSort || ratingFilter {
		bayes = r.bayesianExpr()
	}

	// ORDER BY clause.
	//   rating       → rated first (unrated rt.rcnt IS NULL sinks), then
	//                  the Bayesian score by direction.
	//   release_date → the only other nullable sort col; pin NULLS LAST so
	//                  unknown-date rows don't crowd a "newest first" view.
	orderClause := sortCol + " " + f.SortOrder
	switch {
	case ratingSort:
		orderClause = "(rt.rcnt IS NULL), " + bayes + " " + f.SortOrder
	case sortCol == "g.release_date":
		orderClause += " NULLS LAST"
	}

	type idRow struct {
		ID int `gorm:"column:id"`
	}

	if !hasResourceFilter(f) {
		build := func() *gorm.DB {
			q := r.db.Table("galgame g")
			if ratingSort || ratingFilter {
				q = q.Joins(ratingAggJoin)
			}
			q = applyReleaseFilter(q, f)
			if ratingFilter {
				q = applyRatingFilter(q, f, bayes)
			}
			return q
		}
		build().Select("COUNT(*)").Scan(&total)
		var rows []idRow
		build().
			Select("g.id").
			Order(orderClause).
			Offset((f.Page - 1) * f.Limit).Limit(f.Limit).
			Scan(&rows)
		ids = make([]int, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}
		return
	}

	// Join with galgame_resource and apply filters. The rating filter lives
	// in `inner` (it decides which ids qualify); the rating SORT needs the
	// join on the outer query too (to ORDER by it).
	inner := r.db.Table("galgame g").
		Select("DISTINCT g.id").
		Joins("JOIN galgame_resource gr ON gr.galgame_id = g.id")
	if ratingFilter {
		inner = inner.Joins(ratingAggJoin)
	}

	inner = applyReleaseFilter(inner, f)
	if f.Type != "" && f.Type != "all" {
		inner = inner.Where("gr.type = ?", f.Type)
	}
	if f.Language != "" && f.Language != "all" {
		inner = inner.Where("gr.language = ?", f.Language)
	}
	if f.Platform != "" && f.Platform != "all" {
		inner = inner.Where("gr.platform = ?", f.Platform)
	}
	if len(f.IncludeProviders) > 0 {
		inner = inner.Where("gr.provider && ?", providerArrayLit(f.IncludeProviders))
	}
	if len(f.ExcludeOnlyProviders) > 0 {
		allowed := providersExcluding(f.ExcludeOnlyProviders)
		if len(allowed) > 0 {
			inner = inner.Where("gr.provider && ?", providerArrayLit(allowed))
		}
	}
	if ratingFilter {
		inner = applyRatingFilter(inner, f, bayes)
	}

	r.db.Table("(?) AS sub", inner).Select("COUNT(*)").Scan(&total)

	main := r.db.Table("galgame g").
		Select("g.id").
		Joins("JOIN galgame_resource gr ON gr.galgame_id = g.id")
	groupBy := "g.id, " + sortCol
	if ratingSort {
		main = main.Joins(ratingAggJoin)
		groupBy = "g.id, rt.rsum, rt.rcnt"
	}

	var rows []idRow
	main.
		Where("gr.galgame_id IN (?)", inner).
		Group(groupBy).
		Order(orderClause).
		Offset((f.Page - 1) * f.Limit).Limit(f.Limit).
		Scan(&rows)

	ids = make([]int, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return
}

// bayesianExpr builds the SQL fragment for the galgame Bayesian rating
// average using the LIVE global mean m (cheap: AVG over the small
// galgame_rating table) and the const prior C. m and C are server-owned
// numbers (never user input), so formatting them straight into the
// fragment is injection-safe and avoids pgx param-type inference issues
// in ORDER BY / arithmetic contexts.
//
// References rt.rsum / rt.rcnt from ratingAggJoin; COALESCE handles the
// unrated (LEFT JOIN NULL) case → score = (C·m)/C = m. C>0 guarantees a
// non-zero denominator.
func (r *GalgameListRepository) bayesianExpr() string {
	var m float64
	r.db.Table("galgame_rating").Select("COALESCE(AVG(overall), 0)").Scan(&m)
	c := strconv.FormatFloat(bayesianPriorC, 'f', -1, 64)
	ms := strconv.FormatFloat(m, 'f', 6, 64)
	return "(" + c + " * " + ms + " + COALESCE(rt.rsum, 0)) / (" +
		c + " + COALESCE(rt.rcnt, 0))"
}

// applyRatingFilter adds the advanced rating WHEREs. minRatingCount is a
// high-confidence gate (only games with enough votes); minRating filters
// on the smoothed Bayesian score so a lone 10/10 doesn't slip through.
func applyRatingFilter(q *gorm.DB, f model.GalgameListFilter, bayes string) *gorm.DB {
	if f.MinRatingCount > 0 {
		q = q.Where("COALESCE(rt.rcnt, 0) >= ?", f.MinRatingCount)
	}
	if f.MinRating > 0 {
		q = q.Where(bayes+" >= ?", f.MinRating)
	}
	return q
}

// applyReleaseFilter adds inclusive release_date bounds when present.
// Bounds are pre-resolved "YYYY-MM-DD" strings (PG casts the literal to
// date). Setting either bound drops NULL release_date rows automatically
// — PG evaluates the comparison to UNKNOWN for NULL, excluding them
// (wiki §17.4). Both empty → no-op, returns the query unchanged.
func applyReleaseFilter(q *gorm.DB, f model.GalgameListFilter) *gorm.DB {
	if f.ReleasedFrom != "" {
		q = q.Where("g.release_date >= ?", f.ReleasedFrom)
	}
	if f.ReleasedTo != "" {
		q = q.Where("g.release_date <= ?", f.ReleasedTo)
	}
	// Discontinuous month set (wiki §17.10), AND-combined with the year
	// range. Non-sargable EXTRACT, but it only rechecks the candidate set
	// the release_date btree range scan already narrowed. NULL release_date
	// → EXTRACT(NULL) → NULL → not IN → dropped, consistent with §17.4.
	if len(f.ReleasedMonths) > 0 {
		q = q.Where("EXTRACT(MONTH FROM g.release_date)::int IN ?", f.ReleasedMonths)
	}
	return q
}

func hasResourceFilter(f model.GalgameListFilter) bool {
	return (f.Type != "" && f.Type != "all") ||
		(f.Language != "" && f.Language != "all") ||
		(f.Platform != "" && f.Platform != "all") ||
		len(f.IncludeProviders) > 0 ||
		len(f.ExcludeOnlyProviders) > 0
}

func providerArrayLit(providers []string) string {
	return "{" + strings.Join(providers, ",") + "}"
}

func providersExcluding(excluded []string) []string {
	exSet := map[string]bool{}
	for _, e := range excluded {
		exSet[e] = true
	}
	out := make([]string, 0, len(allProviders))
	for _, p := range allProviders {
		if !exSet[p] {
			out = append(out, p)
		}
	}
	return out
}
