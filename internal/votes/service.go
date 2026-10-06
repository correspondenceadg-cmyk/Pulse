package votes

import (
	"context"
	"errors"
)

var (
	ErrInvalidValue = errors.New("value must be 1, -1, or 0")
	ErrEventMissing = errors.New("event not found")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Vote(ctx context.Context, eventID, userID string, value int) (*Stats, error) {
	if value != -1 && value != 0 && value != 1 {
		return nil, ErrInvalidValue
	}

	exists, err := s.repo.EventExists(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrEventMissing
	}

	switch value {
	case 1, -1:
		if err := s.repo.Upsert(ctx, eventID, userID, int16(value)); err != nil {
			return nil, err
		}
	case 0:
		if err := s.repo.Delete(ctx, eventID, userID); err != nil {
			return nil, err
		}
	}

	return s.repo.Stats(ctx, eventID, userID)
}

func (s *Service) Stats(ctx context.Context, eventID, userID string) (*Stats, error) {
	exists, err := s.repo.EventExists(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrEventMissing
	}
	return s.repo.Stats(ctx, eventID, userID)
}
