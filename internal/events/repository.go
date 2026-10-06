package events

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("event not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type ListFilter struct {
	MinLng   *float64
	MinLat   *float64
	MaxLng   *float64
	MaxLat   *float64
	Category string
	Upcoming bool
	CursorAt *time.Time
	CursorID *string
	Limit    int
}

const selectColumns = `
	id, owner_id, title, coalesce(description,''), category,
	lat, lng, starts_at, ends_at,
	coalesce(venue,''), coalesce(address,''), visibility,
	created_at, updated_at
`

func scanEvent(row pgx.Row) (*Event, error) {
	var e Event
	err := row.Scan(
		&e.ID, &e.OwnerID, &e.Title, &e.Description, &e.Category,
		&e.Lat, &e.Lng, &e.StartsAt, &e.EndsAt,
		&e.Venue, &e.Address, &e.Visibility,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) Create(ctx context.Context, e *Event) (*Event, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO events
			(owner_id, title, description, category, lat, lng, starts_at, ends_at, venue, address, visibility)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+selectColumns,
		e.OwnerID, e.Title, e.Description, e.Category,
		e.Lat, e.Lng, e.StartsAt, e.EndsAt,
		e.Venue, e.Address, e.Visibility,
	)
	created, err := scanEvent(row)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectColumns+`
		FROM events WHERE id = $1
	`, id)
	e, err := scanEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return e, err
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]*Event, error) {
	args := []any{}
	where := []string{"1=1"}

	if f.MinLng != nil && f.MinLat != nil && f.MaxLng != nil && f.MaxLat != nil {
		where = append(where, "lng BETWEEN $"+itoa(len(args)+1)+" AND $"+itoa(len(args)+2))
		args = append(args, *f.MinLng, *f.MaxLng)
		where = append(where, "lat BETWEEN $"+itoa(len(args)+1)+" AND $"+itoa(len(args)+2))
		args = append(args, *f.MinLat, *f.MaxLat)
	}

	if f.Category != "" {
		where = append(where, "category = $"+itoa(len(args)+1))
		args = append(args, f.Category)
	}

	if f.Upcoming {
		where = append(where, "starts_at >= now()")
	}

	if f.CursorAt != nil && f.CursorID != nil {
		where = append(where,
			"(starts_at, id) > ($"+itoa(len(args)+1)+", $"+itoa(len(args)+2)+")")
		args = append(args, *f.CursorAt, *f.CursorID)
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := "SELECT " + selectColumns + " FROM events WHERE "
	q += joinAnd(where)
	q += " ORDER BY starts_at ASC, id ASC LIMIT $" + itoa(len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) Update(ctx context.Context, e *Event) (*Event, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE events SET
			title=$1, description=$2, category=$3,
			lat=$4, lng=$5, starts_at=$6, ends_at=$7,
			venue=$8, address=$9, visibility=$10,
			updated_at=now()
		WHERE id=$11 AND owner_id=$12
		RETURNING `+selectColumns,
		e.Title, e.Description, e.Category,
		e.Lat, e.Lng, e.StartsAt, e.EndsAt,
		e.Venue, e.Address, e.Visibility,
		e.ID, e.OwnerID,
	)
	updated, err := scanEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return updated, err
}

func (r *Repository) Delete(ctx context.Context, id, ownerID string) error {
	cmd, err := r.pool.Exec(ctx,
		`DELETE FROM events WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

func joinAnd(parts []string) string {
	if len(parts) == 0 {
		return "1=1"
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += " AND " + p
	}
	return out
}