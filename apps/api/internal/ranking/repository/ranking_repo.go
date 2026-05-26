package repository

import (
	"gorm.io/gorm"
)

type RankingRepository struct {
	db *gorm.DB
}

func NewRankingRepository(db *gorm.DB) *RankingRepository {
	return &RankingRepository{db: db}
}

// ──────────────────────────────────────────
// Row projections
// ──────────────────────────────────────────

type GalgameLocalRow struct {
	ID    int `gorm:"column:id"`
	Value int `gorm:"column:value"`
}

// TopicRankingRow returns a topic ranking row. Identity is hydrated by the
// service layer via userclient.
type TopicRankingRow struct {
	ID     int    `gorm:"column:id"`
	Title  string `gorm:"column:title"`
	UserID int    `gorm:"column:user_id"`
	Value  int    `gorm:"column:value"`
}

// UserRankingRow returns a user ranking row keyed by user_id. Sorting fields
// (e.g. moemoepoint) live in kungal_user_state. Identity (name/avatar/bio) is
// hydrated by the service layer via userclient.
type UserRankingRow struct {
	UserID int `gorm:"column:user_id"`
	Value  int `gorm:"column:value"`
}

// ──────────────────────────────────────────
// Queries
// ──────────────────────────────────────────

// FindGalgameLocal returns (id, sort_value) pairs from the galgame table
// sorted by the requested field.
func (r *RankingRepository) FindGalgameLocal(sortField, sortOrder string, page, limit int) []GalgameLocalRow {
	var rows []GalgameLocalRow
	r.db.Table("galgame").
		Select("id, "+sortField+" AS value").
		Order(sortField + " " + sortOrder).
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&rows)
	return rows
}

// FindTopicRanking returns topic ranking rows. Identity is hydrated at the
// service layer.
//
// isSFW=true filters out is_nsfw=true rows so anonymous / SEO-bot callers
// can't crawl NSFW topics through the ranking page (and so cookie-off
// users get a clean list). is_nsfw is kungal-local data, so the filter
// is correctly applied at the SQL layer here (unlike galgame.content_limit
// which lives only on wiki).
func (r *RankingRepository) FindTopicRanking(sortField, sortOrder string, page, limit int, isSFW bool) []TopicRankingRow {
	var rows []TopicRankingRow
	q := r.db.Table("topic t").
		Select(`t.id, t.title, t.user_id, t.` + sortField + ` AS value`).
		Where("t.status != 1")
	if isSFW {
		q = q.Where("t.is_nsfw = false")
	}
	q.Order("t." + sortField + " " + sortOrder).
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return rows
}

// FindUserRanking returns user ranking rows ordered by a kungal_user_state
// column (currently only moemoepoint). Identity (name/avatar/bio) is hydrated
// at the service layer via userclient since the user table is no longer the
// source of truth.
func (r *RankingRepository) FindUserRanking(sortField, sortOrder string, page, limit int) []UserRankingRow {
	var rows []UserRankingRow
	r.db.Table("kungal_user_state").
		Select(`user_id, ` + sortField + ` AS value`).
		Order(sortField + " " + sortOrder).
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return rows
}
