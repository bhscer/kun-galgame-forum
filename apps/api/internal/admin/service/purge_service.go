package service

import (
	"kun-galgame-api/internal/admin/dto"
	"kun-galgame-api/internal/admin/repository"
	"kun-galgame-api/pkg/errors"
)

type PurgeService struct {
	repo *repository.PurgeRepository
}

func NewPurgeService(repo *repository.PurgeRepository) *PurgeService {
	return &PurgeService{repo: repo}
}

// GetUserContentStats previews what a purge would remove.
func (s *PurgeService) GetUserContentStats(userID int) dto.UserContentStats {
	return s.repo.CountUserContent(userID)
}

// PurgeUserContent hard-deletes all of the user's kungal content +
// interactions and returns the breakdown of what was removed.
func (s *PurgeService) PurgeUserContent(userID int) (dto.UserContentStats, *errors.AppError) {
	stats, err := s.repo.PurgeUserContent(userID)
	if err != nil {
		return dto.UserContentStats{}, errors.ErrInternal("清除用户内容失败")
	}
	return stats, nil
}
