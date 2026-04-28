package model

type ChatMessageCreateInput struct {
	Text string `json:"text"`
}

type ChatMarkReadInput struct {
	MessageIDs []int64 `json:"message_ids"`
}
