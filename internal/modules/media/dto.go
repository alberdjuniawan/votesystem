package media

import "time"

type MediaResponse struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"created_at"`
}
