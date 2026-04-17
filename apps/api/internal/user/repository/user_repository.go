package repository

import (
	"kun-galgame-api/internal/user/dto"
	"kun-galgame-api/internal/user/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ──────────────────────────────────────────
// Basic CRUD
// ──────────────────────────────────────────

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByName(name string) (*model.User, error) {
	var user model.User
	err := r.db.Where("name = ?", name).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByOAuthSub(sub string) (*model.User, error) {
	var account model.OAuthAccount
	err := r.db.Where("sub = ?", sub).Preload("User").First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account.User, nil
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id int) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) UpdateField(uid int, field string, value any) error {
	return r.db.Model(&model.User{}).Where("id = ?", uid).Update(field, value).Error
}

// ──────────────────────────────────────────
// OAuth
// ──────────────────────────────────────────

func (r *UserRepository) CreateOAuthAccount(account *model.OAuthAccount) error {
	return r.db.Create(account).Error
}

func (r *UserRepository) LinkOAuthAccount(userID int, sub string) error {
	return r.db.Create(&model.OAuthAccount{
		UserID:   userID,
		Provider: "kun-oauth",
		Sub:      sub,
	}).Error
}

// ──────────────────────────────────────────
// User operations
// ──────────────────────────────────────────

func (r *UserRepository) IncrementMoemoepoint(userID int, amount int) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("moemoepoint", gorm.Expr("moemoepoint + ?", amount)).Error
}

func (r *UserRepository) CheckIn(uid int, points int) error {
	return r.db.Model(&model.User{}).Where("id = ?", uid).
		Updates(map[string]any{
			"daily_check_in": 1,
			"moemoepoint":    gorm.Expr("moemoepoint + ?", points),
		}).Error
}

func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Where("LOWER(name) = LOWER(?)", username).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) UpdateUsernameWithCost(uid int, username string, cost int) error {
	return r.db.Model(&model.User{}).Where("id = ?", uid).
		Updates(map[string]any{
			"name":        username,
			"moemoepoint": gorm.Expr("moemoepoint - ?", cost),
		}).Error
}

// ──────────────────────────────────────────
// Stats & counts
// ──────────────────────────────────────────

type UserStats = model.UserStats

func (r *UserRepository) GetUserStats(uid int) (*model.UserStats, error) {
	var stats model.UserStats
	err := r.db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM topic WHERE user_id = @uid) AS topic,
			(SELECT COUNT(*) FROM topic_poll WHERE user_id = @uid) AS topic_poll,
			(SELECT COUNT(*) FROM topic_reply WHERE user_id = @uid) AS reply_created,
			(SELECT COUNT(*) FROM topic_comment WHERE user_id = @uid) AS comment_created,
			(SELECT COUNT(*) FROM galgame_comment WHERE user_id = @uid) AS galgame_comment,
			(SELECT COUNT(*) FROM galgame_rating WHERE user_id = @uid) AS galgame_rating,
			(SELECT COUNT(*) FROM galgame_resource WHERE user_id = @uid) AS galgame_resource,
			(SELECT COUNT(*) FROM galgame_website WHERE user_id = @uid) AS galgame_toolset,
			(SELECT COUNT(*) FROM galgame_toolset_resource WHERE user_id = @uid) AS galgame_toolset_resource,
			(SELECT COUNT(*) FROM topic_upvote WHERE topic_id IN (SELECT id FROM topic WHERE user_id = @uid)) AS upvote,
			(SELECT COUNT(*) FROM topic_like WHERE topic_id IN (SELECT id FROM topic WHERE user_id = @uid)) AS "like",
			(SELECT COUNT(*) FROM topic_dislike WHERE topic_id IN (SELECT id FROM topic WHERE user_id = @uid)) AS dislike,
			(SELECT COUNT(*) FROM topic WHERE user_id = @uid AND created >= CURRENT_DATE) AS daily_topic_count
	`, map[string]any{"uid": uid}).Scan(&stats).Error
	return &stats, err
}

func (r *UserRepository) CountUnreadMessages(uid int) (int64, error) {
	var count int64
	err := r.db.Table("message").
		Where("receiver_id = ? AND status = 'unread'", uid).
		Count(&count).Error
	return count, err
}

func (r *UserRepository) CountUnreadSystemMessages() (int64, error) {
	var count int64
	err := r.db.Table("system_message").
		Where("status = 'unread'").
		Count(&count).Error
	return count, err
}

func (r *UserRepository) CountUnreadChatMessages(uid int) (int64, error) {
	var count int64
	err := r.db.Table("chat_message").
		Where("sender_id != ?", uid).
		Where("chat_room_id IN (SELECT chat_room_id FROM chat_room_participant WHERE user_id = ?)", uid).
		Where("id NOT IN (SELECT chat_message_id FROM chat_message_read_by WHERE user_id = ?)", uid).
		Count(&count).Error
	return count, err
}

// ──────────────────────────────────────────
// User galgames / topics queries
// ──────────────────────────────────────────

func (r *UserRepository) FindUserGalgameIDs(uid int, queryType string, page, limit int) ([]int, int64, error) {
	offset := (page - 1) * limit
	var total int64

	baseQuery := r.db.Table("galgame").Select("galgame.id")

	switch queryType {
	case "galgame_like":
		baseQuery = baseQuery.
			Joins("JOIN galgame_like ON galgame_like.galgame_id = galgame.id").
			Where("galgame_like.user_id = ?", uid)
	case "galgame_favorite":
		baseQuery = baseQuery.
			Joins("JOIN galgame_favorite ON galgame_favorite.galgame_id = galgame.id").
			Where("galgame_favorite.user_id = ?", uid)
	case "galgame_comment":
		baseQuery = baseQuery.
			Joins("JOIN galgame_comment ON galgame_comment.galgame_id = galgame.id").
			Where("galgame_comment.user_id = ?", uid).
			Group("galgame.id")
	case "galgame_comment_target":
		// Comments targeting this user's galgame comments
		baseQuery = baseQuery.
			Joins("JOIN galgame_comment ON galgame_comment.galgame_id = galgame.id").
			Where("galgame_comment.target_user_id = ? AND galgame_comment.user_id != ?", uid, uid).
			Group("galgame.id")
	case "galgame_comment_like":
		baseQuery = baseQuery.
			Joins("JOIN galgame_comment ON galgame_comment.galgame_id = galgame.id").
			Joins("JOIN galgame_comment_like ON galgame_comment_like.galgame_comment_id = galgame_comment.id").
			Where("galgame_comment_like.user_id = ?", uid).
			Group("galgame.id")
	default:
		return []int{}, 0, nil
	}

	baseQuery.Count(&total)

	type idRow struct {
		ID int `gorm:"column:id"`
	}
	var rows []idRow
	err := baseQuery.Order("galgame.created DESC").
		Offset(offset).Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	return ids, total, nil
}

func (r *UserRepository) FindUserTopics(uid int, queryType string, page, limit int) ([]dto.UserTopic, int64, error) {
	offset := (page - 1) * limit
	var results []dto.UserTopic
	var total int64

	baseQuery := r.db.Table("topic").
		Select("topic.id, topic.title, topic.created")

	switch queryType {
	case "topic":
		baseQuery = baseQuery.Where("topic.user_id = ?", uid)
	case "topic_like":
		baseQuery = baseQuery.
			Joins("JOIN topic_like ON topic_like.topic_id = topic.id").
			Where("topic_like.user_id = ?", uid)
	case "topic_upvote":
		baseQuery = baseQuery.
			Joins("JOIN topic_upvote ON topic_upvote.topic_id = topic.id").
			Where("topic_upvote.user_id = ?", uid)
	case "topic_favorite":
		baseQuery = baseQuery.
			Joins("JOIN topic_favorite ON topic_favorite.topic_id = topic.id").
			Where("topic_favorite.user_id = ?", uid)
	case "topic_hide":
		baseQuery = baseQuery.Where("topic.user_id = ? AND topic.status = 1", uid)
	default:
		baseQuery = baseQuery.Where("topic.user_id = ?", uid)
	}

	baseQuery.Count(&total)
	err := baseQuery.Order("topic.created DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

// ──────────────────────────────────────────
// User replies
// ──────────────────────────────────────────

type UserReply struct {
	TopicID int    `gorm:"column:topic_id" json:"topicId"`
	Content string `gorm:"column:content" json:"content"`
	Created string `gorm:"column:created" json:"created"`
}

func (r *UserRepository) FindUserReplies(uid int, queryType string, page, limit int) ([]UserReply, int64, error) {
	offset := (page - 1) * limit
	var results []UserReply
	var total int64

	baseQuery := r.db.Table("topic_reply").
		Select("topic_reply.topic_id, topic_reply.content, topic_reply.created")

	switch queryType {
	case "reply_target":
		baseQuery = baseQuery.
			Where("topic_reply.topic_id IN (SELECT id FROM topic WHERE user_id = ?) AND topic_reply.user_id != ?", uid, uid)
	case "reply_like":
		baseQuery = baseQuery.
			Joins("JOIN topic_reply_like ON topic_reply_like.topic_reply_id = topic_reply.id").
			Where("topic_reply_like.user_id = ?", uid)
	default: // reply_created
		baseQuery = baseQuery.Where("topic_reply.user_id = ?", uid)
	}

	baseQuery.Count(&total)
	err := baseQuery.Order("topic_reply.created DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

// ──────────────────────────────────────────
// User comments
// ──────────────────────────────────────────

type UserComment struct {
	TopicID int    `gorm:"column:topic_id" json:"topicId"`
	Content string `gorm:"column:content" json:"content"`
	Created string `gorm:"column:created" json:"created"`
}

func (r *UserRepository) FindUserComments(uid int, queryType string, page, limit int) ([]UserComment, int64, error) {
	offset := (page - 1) * limit
	var results []UserComment
	var total int64

	baseQuery := r.db.Table("topic_comment").
		Select("topic_comment.topic_id, topic_comment.content, topic_comment.created")

	switch queryType {
	case "comment_target":
		baseQuery = baseQuery.
			Where("topic_comment.target_user_id = ? AND topic_comment.user_id != ?", uid, uid)
	case "comment_like":
		baseQuery = baseQuery.
			Joins("JOIN topic_comment_like ON topic_comment_like.topic_comment_id = topic_comment.id").
			Where("topic_comment_like.user_id = ?", uid)
	default: // comment_created
		baseQuery = baseQuery.Where("topic_comment.user_id = ?", uid)
	}

	baseQuery.Count(&total)
	err := baseQuery.Order("topic_comment.created DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

// ──────────────────────────────────────────
// User galgame resources
// ──────────────────────────────────────────

type UserResource struct {
	ID        int    `gorm:"column:id" json:"id"`
	GalgameID int    `gorm:"column:galgame_id" json:"galgameId"`
	Type      string `gorm:"column:type" json:"type"`
	Language  string `gorm:"column:language" json:"language"`
	Platform  string `gorm:"column:platform" json:"platform"`
	Size      string `gorm:"column:size" json:"size"`
	Code      string `gorm:"column:code" json:"code"`
	Password  string `gorm:"column:password" json:"password"`
	Note      string `gorm:"column:note" json:"note"`
	Status    int    `gorm:"column:status" json:"status"`
	Created   string `gorm:"column:created" json:"created"`
}

type ResourceLink struct {
	ResourceID int    `gorm:"column:galgame_resource_id"`
	URL        string `gorm:"column:url"`
}

func (r *UserRepository) FindUserResources(uid int, queryType string, page, limit int) ([]UserResource, int64, error) {
	offset := (page - 1) * limit
	var results []UserResource
	var total int64

	baseQuery := r.db.Table("galgame_resource").
		Select("galgame_resource.id, galgame_resource.galgame_id, galgame_resource.type, galgame_resource.language, galgame_resource.platform, galgame_resource.size, galgame_resource.code, galgame_resource.password, galgame_resource.note, galgame_resource.status, galgame_resource.created")

	switch queryType {
	case "expire":
		baseQuery = baseQuery.Where("galgame_resource.user_id = ? AND galgame_resource.status = 1", uid)
	case "galgame_resource_like":
		baseQuery = baseQuery.
			Joins("JOIN galgame_resource_like ON galgame_resource_like.galgame_resource_id = galgame_resource.id").
			Where("galgame_resource_like.user_id = ?", uid)
	default: // valid
		baseQuery = baseQuery.Where("galgame_resource.user_id = ? AND galgame_resource.status = 0", uid)
	}

	baseQuery.Count(&total)
	err := baseQuery.Order("galgame_resource.created DESC").Offset(offset).Limit(limit).Scan(&results).Error
	return results, total, err
}

func (r *UserRepository) FindResourceLinks(resourceIDs []int) (map[int][]string, error) {
	var links []ResourceLink
	err := r.db.Table("galgame_resource_link").
		Select("galgame_resource_id, url").
		Where("galgame_resource_id IN ?", resourceIDs).
		Scan(&links).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int][]string)
	for _, l := range links {
		result[l.ResourceID] = append(result[l.ResourceID], l.URL)
	}
	return result, nil
}

// ──────────────────────────────────────────
// User galgame ratings
// ──────────────────────────────────────────

type UserRating struct {
	ID           int    `gorm:"column:id" json:"id"`
	GalgameID    int    `gorm:"column:galgame_id" json:"galgameId"`
	Recommend    string `gorm:"column:recommend" json:"recommend"`
	Overall      int    `gorm:"column:overall" json:"overall"`
	View         int    `gorm:"column:view" json:"view"`
	Art          int    `gorm:"column:art" json:"art"`
	Story        int    `gorm:"column:story" json:"story"`
	Music        int    `gorm:"column:music" json:"music"`
	Character    int    `gorm:"column:character" json:"character"`
	Route        int    `gorm:"column:route" json:"route"`
	System       int    `gorm:"column:system" json:"system"`
	Voice        int    `gorm:"column:voice" json:"voice"`
	ReplayValue  int    `gorm:"column:replay_value" json:"replay_value"`
	GalgameType  string `gorm:"column:galgame_type" json:"-"` // raw JSON
	PlayStatus   string `gorm:"column:play_status" json:"play_status"`
	SpoilerLevel string `gorm:"column:spoiler_level" json:"spoiler_level"`
	LikeCount    int    `gorm:"column:like_count" json:"likeCount"`
	UserID       int    `gorm:"column:user_id" json:"-"`
	UserName     string `gorm:"column:user_name" json:"-"`
	UserAvatar   string `gorm:"column:user_avatar" json:"-"`
	Created      string `gorm:"column:created" json:"created"`
	Updated      string `gorm:"column:updated" json:"updated"`
}

func (r *UserRepository) FindUserRatings(uid int, page, limit int) ([]UserRating, int64, error) {
	offset := (page - 1) * limit
	var results []UserRating
	var total int64

	r.db.Table("galgame_rating").Where("user_id = ?", uid).Count(&total)

	err := r.db.Table("galgame_rating").
		Select(`galgame_rating.id, galgame_rating.galgame_id, galgame_rating.recommend, galgame_rating.overall, galgame_rating.view,
			galgame_rating.art, galgame_rating.story, galgame_rating.music, galgame_rating.character, galgame_rating.route, galgame_rating.system, galgame_rating.voice, galgame_rating.replay_value,
			galgame_rating.galgame_type, galgame_rating.play_status, galgame_rating.spoiler_level, galgame_rating.like_count,
			galgame_rating.user_id, u.name AS user_name, u.avatar AS user_avatar,
			galgame_rating.created, galgame_rating.updated`).
		Joins(`LEFT JOIN "user" u ON u.id = galgame_rating.user_id`).
		Where("galgame_rating.user_id = ?", uid).
		Order("galgame_rating.created DESC").Offset(offset).Limit(limit).
		Scan(&results).Error
	return results, total, err
}
