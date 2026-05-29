package service

import (
	"context"
	"math/rand/v2"

	galgameClient "kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/internal/user/dto"
	"kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"

	"github.com/redis/go-redis/v9"
)

// UserService surfaces the kungal-side user pages: profile, status,
// floating card, and check-in. Identity (name / avatar / bio / status /
// roles) is fetched from OAuth via userclient. Per-site state
// (moemoepoint, daily counters) lives in kungal_user_state. Content
// counts (topic / reply / etc.) come from local aggregates.
type UserService struct {
	stateRepo     *repository.StateRepository
	userStatsRepo *repository.UserStatsRepository
	rdb           *redis.Client
	wikiClient    *galgameClient.GalgameClient
	userClient    *userclient.Client
}

func NewUserService(
	stateRepo *repository.StateRepository,
	userStatsRepo *repository.UserStatsRepository,
	rdb *redis.Client,
	wikiClient *galgameClient.GalgameClient,
	userClient *userclient.Client,
) *UserService {
	return &UserService{
		stateRepo:     stateRepo,
		userStatsRepo: userStatsRepo,
		rdb:           rdb,
		wikiClient:    wikiClient,
		userClient:    userClient,
	}
}

// ──────────────────────────────────────────
// Profile
// ──────────────────────────────────────────

func (s *UserService) GetUserProfile(ctx context.Context, userID int) (*dto.UserProfileDetail, *errors.AppError) {
	u, ok, err := s.userClient.User(ctx, userID)
	if err != nil {
		return nil, errors.ErrInternal("查询用户信息失败")
	}
	if !ok {
		return nil, errors.ErrNotFound("未找到该用户")
	}
	// Banned users surface only id+name so callers can render a "已封禁"
	// placeholder; we don't leak avatar / bio / stats.
	if u.Status != 0 {
		return &dto.UserProfileDetail{ID: u.ID, Name: u.Name, Status: u.Status}, nil
	}

	stats, err := s.userStatsRepo.GetUserStats(userID)
	if err != nil {
		return nil, errors.ErrInternal("获取用户统计失败")
	}
	state, _ := s.stateRepo.FindByID(userID)
	moe := 0
	if state != nil {
		moe = state.Moemoepoint
	}

	profile := &dto.UserProfileDetail{
		ID:          u.ID,
		Name:        u.Name,
		Avatar:      u.Avatar,
		Role:        middleware.RoleFromOAuthRoles(u.Roles),
		Status:      u.Status,
		Moemoepoint: moe,
		Bio:         u.Bio,
	}
	// CreatedAt: kungal_user_state.created marks "first seen on this
	// site" — close enough for a community profile.
	if state != nil {
		profile.CreatedAt = state.CreatedAt
	}

	profile.Topic = stats.Topic
	profile.TopicPoll = stats.TopicPoll
	profile.ReplyCreated = stats.ReplyCreated
	profile.CommentCreated = stats.CommentCreated
	profile.GalgameComment = stats.GalgameComment
	profile.GalgameRating = stats.GalgameRating
	profile.GalgameResource = stats.GalgameResource
	profile.GalgameToolset = stats.GalgameToolset
	profile.GalgameToolsetResource = stats.GalgameToolsetResource
	profile.Upvote = stats.Upvote
	profile.Like = stats.Like
	profile.Dislike = stats.Dislike
	profile.DailyTopicCount = stats.DailyTopicCount

	if wikiStats, err := s.wikiClient.GetUserStats(ctx, userID); err == nil && wikiStats != nil {
		profile.Galgame = wikiStats.GalgameCreated
		profile.DailyGalgameCount = wikiStats.GalgameCreatedToday
		profile.ContributeGalgame = wikiStats.GalgameContributed
	}

	return profile, nil
}

// ──────────────────────────────────────────
// Check-in / status
// ──────────────────────────────────────────

func (s *UserService) CheckIn(ctx context.Context, userID int) (int, *errors.AppError) {
	// Atomic once-per-day gate: CheckIn only applies when daily_check_in = 0
	// (reset at calendar midnight by the daily cron). No read-then-write race
	// and no external rate limiter — a repeat attempt today simply applies
	// nothing and we report "已签到".
	points := rand.IntN(8) // 0-7
	applied, err := s.stateRepo.CheckIn(userID, points)
	if err != nil {
		return 0, errors.ErrInternal("签到失败")
	}
	if !applied {
		return 0, errors.ErrBadRequest("您今天已经签到过了")
	}
	return points, nil
}

func (s *UserService) GetUserStatus(ctx context.Context, userID int) (*dto.UserStatusResponse, *errors.AppError) {
	// A freshly-registered user may not have a state row yet (the
	// callback flow creates it lazily). Treat that case as zero-state
	// rather than 404 — the FE Nav.vue otherwise silently drops the
	// moemoepoint chip + check-in badge with no fallback, so brand-new
	// users see a half-broken nav until they click "签到" the first
	// time (which is itself gated behind reading moemoepoint…).
	moe := 0
	isCheckIn := false
	if state, err := s.stateRepo.FindByID(userID); err == nil && state != nil {
		moe = state.Moemoepoint
		isCheckIn = state.DailyCheckIn == 1
	}

	unreadMessage, _ := s.userStatsRepo.CountUnreadMessages(userID)
	unreadSystem, _ := s.userStatsRepo.CountUnreadSystemMessages(userID)
	unreadChat, _ := s.userStatsRepo.CountUnreadChatMessages(userID)

	return &dto.UserStatusResponse{
		Moemoepoints:  moe,
		IsCheckIn:     isCheckIn,
		HasNewMessage: (unreadMessage + unreadSystem + unreadChat) > 0,
	}, nil
}

// ──────────────────────────────────────────
// Floating hover card
// ──────────────────────────────────────────

func (s *UserService) GetFloatingCard(ctx context.Context, userID int) (*dto.FloatingCardResponse, *errors.AppError) {
	u, ok, err := s.userClient.User(ctx, userID)
	if err != nil {
		return nil, errors.ErrInternal("查询用户信息失败")
	}
	if !ok || u.Status != 0 {
		// Banned or missing — kungal hides the card per agreed policy.
		return nil, errors.ErrNotFound("未找到该用户")
	}

	state, _ := s.stateRepo.FindByID(userID)
	moe := 0
	if state != nil {
		moe = state.Moemoepoint
	}

	stats := s.userStatsRepo.FindFloatingStats(userID)
	return &dto.FloatingCardResponse{
		ID:                   u.ID,
		Name:                 u.Name,
		Avatar:               u.Avatar,
		Moemoepoint:          moe,
		TopicCount:           stats.TopicCount,
		TopicReplyCount:      stats.TopicReplyCount,
		TopicCommentCount:    stats.TopicCommentCount,
		GalgameResourceCount: stats.ResourceCount,
	}, nil
}
