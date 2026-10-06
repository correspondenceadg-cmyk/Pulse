package events

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrInvalid   = errors.New("invalid input")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Title       string
	Description string
	Category    string
	Lat         float64
	Lng         float64
	StartsAt    time.Time
	EndsAt      *time.Time
	Venue       string
	Address     string
	Visibility  string
}

func (s *Service) Create(ctx context.Context, ownerID string, in CreateInput) (*Event, error) {
	if err := validate(in); err != nil {
		return nil, err
	}
	e := &Event{
		OwnerID:     ownerID,
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		Category:    in.Category,
		Lat:         in.Lat,
		Lng:         in.Lng,
		StartsAt:    in.StartsAt,
		EndsAt:      in.EndsAt,
		Venue:       in.Venue,
		Address:     in.Address,
		Visibility:  in.Visibility,
	}
	return s.repo.Create(ctx, e)
}

func (s *Service) Get(ctx context.Context, id, requesterID string) (*Event, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.Visibility == "PRIVATE" && e.OwnerID != requesterID {
		return nil, ErrNotFound
	}
	return e, nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*Event, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) Update(ctx context.Context, id, requesterID string, in CreateInput) (*Event, error) {
	if err := validate(in); err != nil {
		return nil, err
	}
	e := &Event{
		ID:          id,
		OwnerID:     requesterID,
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		Category:    in.Category,
		Lat:         in.Lat,
		Lng:         in.Lng,
		StartsAt:    in.StartsAt,
		EndsAt:      in.EndsAt,
		Venue:       in.Venue,
		Address:     in.Address,
		Visibility:  in.Visibility,
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) Delete(ctx context.Context, id, requesterID string) error {
	return s.repo.Delete(ctx, id, requesterID)
}

func validate(in CreateInput) error {
	if strings.TrimSpace(in.Title) == "" || len(in.Title) > 200 {
		return ErrInvalid
	}
	if in.Category == "" || !validCategories[in.Category] {
		return ErrInvalid
	}
	if in.Visibility == "" || !validVisibilities[in.Visibility] {
		return ErrInvalid
	}
	if in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 {
		return ErrInvalid
	}
	if in.StartsAt.IsZero() {
		return ErrInvalid
	}
	if in.EndsAt != nil && in.EndsAt.Before(in.StartsAt) {
		return ErrInvalid
	}
	return nil
}