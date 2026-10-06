package votes

type Stats struct {
	Up       int    `json:"up"`
	Down     int    `json:"down"`
	Score    int    `json:"score"`
	UserVote *int16 `json:"userVote,omitempty"`
}
