package polls

import "time"

type Option struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Position int    `json:"position"`
}

type Poll struct {
	ID        string     `json:"id"`
	EventID   string     `json:"eventId"`
	Question  string     `json:"question"`
	Status    string     `json:"status"`
	OpensAt   *time.Time `json:"opensAt,omitempty"`
	ClosesAt  *time.Time `json:"closesAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	Options   []Option   `json:"options,omitempty"`
}

type Results struct {
	PollID  string         `json:"pollId"`
	Status  string         `json:"status"`
	Total   int            `json:"total"`
	Counts  map[string]int `json:"counts"`
	UserOpt string         `json:"userOptionId,omitempty"`
}
