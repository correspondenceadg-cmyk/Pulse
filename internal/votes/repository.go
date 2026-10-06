package votes

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Upsert(ctx context.Context, eventID, userID string, value int16) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO event_votes (event_id, user_id, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, user_id)
		DO UPDATE SET value = EXCLUDED.value, created_at = now()
	`, eventID, userID, value)
	return err
}

func (r *Repository) Delete(ctx context.Context, eventID, userID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM event_votes WHERE event_id = $1 AND user_id = $2
	`, eventID, userID)
	return err
}

func (r *Repository) Stats(ctx context.Context, eventID, userID string) (*Stats, error) {
	s := &Stats{}
	err := r.pool.QueryRow(ctx, `
		SELECT
			coalesce(sum(case when value = 1 then 1 else 0 end), 0)::int,
			coalesce(sum(case when value = -1 then 1 else 0 end), 0)::int
		FROM event_votes
		WHERE event_id = $1
	`, eventID).Scan(&s.Up, &s.Down)
	if err != nil {
		return nil, err
	}
	s.Score = s.Up - s.Down

	if userID != "" {
		var v int16
		err := r.pool.QueryRow(ctx, `
			SELECT value FROM event_votes WHERE event_id = $1 AND user_id = $2
		`, eventID, userID).Scan(&v)
		if err == nil {
			s.UserVote = &v
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	return s, nil
}

func (r *Repository) EventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`, eventID).Scan(&exists)
	return exists, err
}
