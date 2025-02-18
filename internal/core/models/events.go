package models

type BotMessage struct {
	Type      string     `json:"type"`
	Payload   BotPayload `json:"payload"`
	Timestamp int64      `json:"timestamp"`
}

type BotMessageType string

const (
	RUN  BotMessageType = "run"
	STOP BotMessageType = "stop"
)

type BotPayload struct {
	BotID       int64  `json:"bot_id"`
	ProjectID   int64  `json:"project_id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	ApiToken    string `json:"api_token"`
}
