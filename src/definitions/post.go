package definitions

import "time"

type Post struct {
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	TextPreview string  `json:"textPreview"`
	Tags       []string `json:"tags"`
	Type       string   `json:"type"`
	Post       string   `json:"post"`
	FirstImage string   `json:"firstImage,omitempty"`
	PostDate   time.Time `json:"postDate"`
}
