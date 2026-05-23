package definitions

import "time"

type Post struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	TextPreview string    `json:"textPreview"`
	Type        string    `json:"type"`
	Post        string    `json:"post"`
	PostDate    time.Time `json:"postDate"`
}
