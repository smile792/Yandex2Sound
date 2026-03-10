package models

type Track struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Artists    string `json:"artists"`
	Album      string `json:"album"`
	DurationMs int    `json:"duration_ms"`
	CoverURL   string `json:"cover_url"`
}

type Playlist struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	TrackCount int     `json:"track_count"`
	CoverURL   string  `json:"cover_url"`
	Tracks     []Track `json:"tracks,omitempty"`
}

type TransferJob struct {
	ID          string        `json:"id"`
	Status      string        `json:"status"`
	Total       int           `json:"total"`
	Current     int           `json:"current"`
	Transferred int           `json:"transferred"`
	NotFound    int           `json:"not_found"`
	Errors      int           `json:"errors"`
	Log         []TransferLog `json:"log"`
	ResultURL   string        `json:"result_url"`
	LastTrack   string        `json:"last_track"`
}

type TransferLog struct {
	TrackTitle string `json:"track_title"`
	Status     string `json:"status"`
}
