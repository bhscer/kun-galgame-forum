package repository

import (
	"strings"

	"kun-galgame-api/internal/galgame/model"

	"gorm.io/gorm"
)

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

	// release_date is the only nullable sort column. PG's default places
	// NULLs FIRST on DESC — i.e. unknown-date rows would crowd the top of
	// a "newest first" view. Pin NULLS LAST so unknowns always sink to the
	// bottom regardless of direction. (With a released_from/to filter set,
	// NULL rows are already excluded, so this only affects the unfiltered
	// release_date sort.)
	orderClause := sortCol + " " + f.SortOrder
	if sortCol == "g.release_date" {
		orderClause += " NULLS LAST"
	}

	if !hasResourceFilter(f) {
		countQ := applyReleaseFilter(r.db.Table("galgame g"), f)
		countQ.Select("COUNT(*)").Scan(&total)
		type idRow struct {
			ID int `gorm:"column:id"`
		}
		var rows []idRow
		applyReleaseFilter(r.db.Table("galgame g"), f).
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

	// Join with galgame_resource and apply filters
	inner := r.db.Table("galgame g").
		Select("DISTINCT g.id").
		Joins("JOIN galgame_resource gr ON gr.galgame_id = g.id")

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

	r.db.Table("(?) AS sub", inner).Select("COUNT(*)").Scan(&total)

	type idRow struct {
		ID int `gorm:"column:id"`
	}
	var rows []idRow
	r.db.Table("galgame g").
		Select("g.id").
		Joins("JOIN galgame_resource gr ON gr.galgame_id = g.id").
		Where("gr.galgame_id IN (?)", inner).
		Group("g.id, " + sortCol).
		Order(orderClause).
		Offset((f.Page - 1) * f.Limit).Limit(f.Limit).
		Scan(&rows)

	ids = make([]int, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return
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
