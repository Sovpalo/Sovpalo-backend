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

type IdeaGenerateInput struct {
	Topic       string   `json:"topic"`
	Context     *string  `json:"context,omitempty"`
	Audience    *string  `json:"audience,omitempty"`
	Constraints []string `json:"constraints,omitempty"`
	Tone        *string  `json:"tone,omitempty"`
	Count       int      `json:"count,omitempty"`
}

type GeneratedIdeaDraft struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source"`
	LLMPrompt   string `json:"llm_prompt"`
}
