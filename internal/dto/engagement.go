package dto

type EngagementQuery struct {
	From  string `form:"from"`
	To    string `form:"to"`
	Page  int    `form:"page,default=1"`
	Limit int    `form:"limit,default=50"`
}

type FunnelStage struct {
	Step        string  `json:"step"`
	Sessions    int64   `json:"sessions"`
	DropOffRate float64 `json:"drop_off_rate"`
}

type FunnelResponse struct {
	Funnel []FunnelStage `json:"funnel"`
}
