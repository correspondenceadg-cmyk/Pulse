package events

import "time"

type Event struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"ownerId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	StartsAt    time.Time  `json:"startsAt"`
	EndsAt      *time.Time `json:"endsAt,omitempty"`
	Venue       string     `json:"venue"`
	Address     string     `json:"address"`
	Visibility  string     `json:"visibility"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

var validCategories = map[string]bool{
	"MUSIC":      true,
	"FOOD":       true,
	"SPORTS":     true,
	"ARTS":       true,
	"TECH":       true,
	"COMMUNITY":  true,
	"OTHER":      true,
}

var validVisibilities = map[string]bool{
	"PUBLIC":  true,
	"PRIVATE": true,
}