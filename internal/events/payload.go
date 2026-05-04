package events

import "time"

type EventPayload struct {
	ID         string                 `json:"id"`
	CampaignID string                 `json:"campaign_id"`
	EventType  string                 `json:"event_type"`
	SourceURL  string                 `json:"source_url"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	Metadata   map[string]interface{} `json:"metadata"`
	OccurredAt time.Time              `json:"occurred_at"`
	SessionID  string                 `json:"session_id"`
	Step       string                 `json:"step"`
}
