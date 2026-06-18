package service

import (
	"context"
	"encoding/json"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/repository"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"
)

// Forum creator-eligibility thresholds. This is the forum's OWN policy — change
// freely here; OAuth and the wiki are untouched (the role + queue live in
// OAuth, the contribution data in the wiki). A user may apply if ANY criterion
// is met. See docs/auth/01-creator-role-design.md.
const (
	creatorMinMergedPRs = 5   // 合并的 PR 数（数据源:wiki /user/:id/stats）
	creatorMinGalgames  = 10  // 已发布 galgame 数（数据源:wiki）
	creatorMinReviews   = 5   // 简评(≥100 字)数（数据源:本论坛 galgame_rating）
	creatorReviewMinLen = 100 // 简评字数门槛
	creatorSource       = "forum"
)

// CreatorEligibility is the forum-side eligibility snapshot (current counts vs
// thresholds) — drives the "离创作者还差…" UI and gates Apply.
type CreatorEligibility struct {
	Eligible          bool  `json:"eligible"`
	MergedPRs         int64 `json:"merged_prs"`
	GalgamesPublished int64 `json:"galgames_published"`
	Reviews100        int64 `json:"reviews_100"`
	NeedMergedPRs     int   `json:"need_merged_prs"`
	NeedGalgames      int   `json:"need_galgames"`
	NeedReviews       int   `json:"need_reviews"`
}

// CreatorService computes forum-side creator eligibility and proxies the
// application to the central OAuth queue. The forum owns the POLICY; OAuth owns
// the queue + admin review + role grant.
type CreatorService struct {
	ratingRepo *repository.RatingRepository
	wikiClient *client.GalgameClient
	userClient *userclient.Client
}

func NewCreatorService(ratingRepo *repository.RatingRepository, wikiClient *client.GalgameClient, userClient *userclient.Client) *CreatorService {
	return &CreatorService{ratingRepo: ratingRepo, wikiClient: wikiClient, userClient: userClient}
}

func (s *CreatorService) eligibility(ctx context.Context, userID int) (*CreatorEligibility, *errors.AppError) {
	stats, err := s.wikiClient.GetUserStats(ctx, userID)
	if err != nil {
		return nil, errors.ErrInternal("获取贡献统计失败")
	}
	reviews, rErr := s.ratingRepo.CountReviewsWithMinLength(userID, creatorReviewMinLen)
	if rErr != nil {
		return nil, errors.ErrInternal("统计简评失败")
	}
	e := &CreatorEligibility{
		MergedPRs:         stats.PRMerged,
		GalgamesPublished: stats.GalgameCreated,
		Reviews100:        reviews,
		NeedMergedPRs:     creatorMinMergedPRs,
		NeedGalgames:      creatorMinGalgames,
		NeedReviews:       creatorMinReviews,
	}
	e.Eligible = e.MergedPRs >= creatorMinMergedPRs ||
		e.GalgamesPublished >= creatorMinGalgames ||
		e.Reviews100 >= creatorMinReviews
	return e, nil
}

// Status returns the user's eligibility snapshot + current OAuth application
// (nil if never applied).
func (s *CreatorService) Status(ctx context.Context, userID int, token string) (*CreatorEligibility, *userclient.CreatorApplication, *errors.AppError) {
	e, appErr := s.eligibility(ctx, userID)
	if appErr != nil {
		return nil, nil, appErr
	}
	app, err := s.userClient.GetMyCreatorApplication(ctx, token)
	if err != nil {
		return nil, nil, errors.ErrInternal("获取申请状态失败")
	}
	return e, app, nil
}

// Apply enforces the forum eligibility gate, then files the application on the
// central OAuth queue with evidence. OAuth's own guards (already-creator /
// one-pending / cooldown) surface via the returned message.
func (s *CreatorService) Apply(ctx context.Context, userID int, token, message string) (*userclient.CreatorApplication, *errors.AppError) {
	e, appErr := s.eligibility(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}
	if !e.Eligible {
		return nil, errors.ErrForbidden("尚不满足创作者申请条件")
	}
	evidence, _ := json.Marshal(map[string]any{
		"merged_prs":         e.MergedPRs,
		"galgames_published": e.GalgamesPublished,
		"reviews_100":        e.Reviews100,
	})
	app, err := s.userClient.CreateCreatorApplication(ctx, token, creatorSource, evidence, message)
	if err != nil {
		if oe, ok := err.(*userclient.OAuthError); ok {
			return nil, errors.ErrBadRequest(oe.Message)
		}
		return nil, errors.ErrInternal("提交申请失败")
	}
	return app, nil
}
