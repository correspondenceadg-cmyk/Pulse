package polls

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalid   = errors.New("invalid input")
	ErrForbidden = errors.New("forbidden")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Question string
	Options  []string
}

func (s *Service) Create(ctx context.Context, eventID, requesterID string, in CreateInput) (*Poll, error) {
	owner, err := s.repo.EventOwner(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if owner != requesterID {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(in.Question) == "" || len(in.Question) > 300 {
		return nil, ErrInvalid
	}
	if len(in.Options) < 2 || len(in.Options) > 10 {
		return nil, ErrInvalid
	}
	labels := make([]string, 0, len(in.Options))
	for _, o := range in.Options {
		o = strings.TrimSpace(o)
		if o == "" || len(o) > 120 {
			return nil, ErrInvalid
		}
		labels = append(labels, o)
	}
	return s.repo.Create(ctx, eventID, in.Question, labels)
}

func (s *Service) ListForEvent(ctx context.Context, eventID string) ([]*Poll, error) {
	return s.repo.ListForEvent(ctx, eventID)
}

func (s *Service) Open(ctx context.Context, pollID, requesterID string) error {
	p, err := s.repo.Get(ctx, pollID)
	if err != nil {
		return err
	}
	owner, err := s.repo.EventOwner(ctx, p.EventID)
	if err != nil {
		return err
	}
	if owner != requesterID {
		return ErrForbidden
	}
	return s.repo.SetStatus(ctx, pollID, "OPEN")
}

func (s *Service) Close(ctx context.Context, pollID, requesterID string) error {
	p, err := s.repo.Get(ctx, pollID)
	if err != nil {
		return err
	}
	owner, err := s.repo.EventOwner(ctx, p.EventID)
	if err != nil {
		return err
	}
	if owner != requesterID {
		return ErrForbidden
	}
	return s.repo.SetStatus(ctx, pollID, "CLOSED")
}

func (s *Service) Vote(ctx context.Context, pollID, optionID, userID string) (*Results, error) {
	p, err := s.repo.Get(ctx, pollID)
	if err != nil {
		return nil, err
	}
	if p.Status != "OPEN" {
		return nil, ErrNotOpen
	}
	if err := s.repo.Vote(ctx, pollID, optionID, userID); err != nil {
		return nil, err
	}
	return s.repo.Results(ctx, pollID, userID)
}

func (s *Service) Results(ctx context.Context, pollID, userID string) (*Results, error) {
	return s.repo.Results(ctx, pollID, userID)
}
