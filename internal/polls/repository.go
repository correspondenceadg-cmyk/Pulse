package polls

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound    = errors.New("poll not found")
	ErrNotOwner    = errors.New("not poll owner")
	ErrNotOpen     = errors.New("poll not open")
	ErrInvalidOpt  = errors.New("invalid option")
	ErrAlreadyVote = errors.New("already voted")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) EventOwner(ctx context.Context, eventID string) (string, error) {
	var owner string
	err := r.pool.QueryRow(ctx,
		`SELECT owner_id FROM events WHERE id = $1`, eventID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return owner, err
}

func (r *Repository) Create(ctx context.Context, eventID, question string, labels []string) (*Poll, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var p Poll
	err = tx.QueryRow(ctx, `
		INSERT INTO polls (event_id, question, status)
		VALUES ($1, $2, 'DRAFT')
		RETURNING id, event_id, question, status, opens_at, closes_at, created_at
	`, eventID, question).Scan(&p.ID, &p.EventID, &p.Question, &p.Status, &p.OpensAt, &p.ClosesAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	for i, label := range labels {
		var opt Option
		err = tx.QueryRow(ctx, `
			INSERT INTO poll_options (poll_id, label, position)
			VALUES ($1, $2, $3)
			RETURNING id, label, position
		`, p.ID, label, i).Scan(&opt.ID, &opt.Label, &opt.Position)
		if err != nil {
			return nil, err
		}
		p.Options = append(p.Options, opt)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListForEvent(ctx context.Context, eventID string) ([]*Poll, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, event_id, question, status, opens_at, closes_at, created_at
		FROM polls WHERE event_id = $1
		ORDER BY created_at DESC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Poll
	for rows.Next() {
		var p Poll
		if err := rows.Scan(&p.ID, &p.EventID, &p.Question, &p.Status, &p.OpensAt, &p.ClosesAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, p := range out {
		opts, err := r.optionsFor(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Options = opts
	}
	return out, nil
}

func (r *Repository) Get(ctx context.Context, pollID string) (*Poll, error) {
	var p Poll
	err := r.pool.QueryRow(ctx, `
		SELECT id, event_id, question, status, opens_at, closes_at, created_at
		FROM polls WHERE id = $1
	`, pollID).Scan(&p.ID, &p.EventID, &p.Question, &p.Status, &p.OpensAt, &p.ClosesAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	opts, err := r.optionsFor(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Options = opts
	return &p, nil
}

func (r *Repository) optionsFor(ctx context.Context, pollID string) ([]Option, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label, position FROM poll_options
		WHERE poll_id = $1 ORDER BY position ASC
	`, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Option
	for rows.Next() {
		var o Option
		if err := rows.Scan(&o.ID, &o.Label, &o.Position); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) SetStatus(ctx context.Context, pollID, status string) error {
	cmd, err := r.pool.Exec(ctx,
		`UPDATE polls SET status = $1 WHERE id = $2`, status, pollID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Vote(ctx context.Context, pollID, optionID, userID string) error {
	// Verify option belongs to poll
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM poll_options WHERE id = $1 AND poll_id = $2)
	`, optionID, pollID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidOpt
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO poll_votes (poll_id, option_id, user_id)
		VALUES ($1, $2, $3)
	`, pollID, optionID, userID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyVote
		}
		return err
	}
	return nil
}

func (r *Repository) Results(ctx context.Context, pollID, userID string) (*Results, error) {
	p, err := r.Get(ctx, pollID)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT option_id, count(*)::int FROM poll_votes
		WHERE poll_id = $1 GROUP BY option_id
	`, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	total := 0
	for rows.Next() {
		var optID string
		var n int
		if err := rows.Scan(&optID, &n); err != nil {
			return nil, err
		}
		counts[optID] = n
		total += n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	r2 := &Results{PollID: pollID, Status: p.Status, Total: total, Counts: counts}

	if userID != "" {
		var optID string
		err := r.pool.QueryRow(ctx,
			`SELECT option_id FROM poll_votes WHERE poll_id = $1 AND user_id = $2`,
			pollID, userID).Scan(&optID)
		if err == nil {
			r2.UserOpt = optID
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	return r2, nil
}

func isUniqueViolation(err error) bool {
	type pgErr interface{ SQLState() string }
	var p pgErr
	if errors.As(err, &p) {
		return p.SQLState() == "23505"
	}
	return false
}
