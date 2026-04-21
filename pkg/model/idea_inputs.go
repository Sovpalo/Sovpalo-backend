package model

type IdeaCreateInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
}

type IdeaUpdateInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
}
