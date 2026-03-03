package models

type Event struct {
	CreatedAt string `json:"createdAt"`
	Type      string `json:"type"`
	DID       string `json:"did"`
	Path      string `json:"path"`
}

type Post struct {
	Handle    string   `json:"handle"`
	Content   string   `json:"content"`
	Likes     int      `json:"likes"`
	Quotes    int      `json:"quotes"`
	Replies   int      `json:"replies"`
	Reposts   int      `json:"reposts"`
	ImageUrls []string `json:"imageUrls"`
	Parent    *Post    `json:"parent,omitempty"`
}
