package model

type IdeaCreateInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}
