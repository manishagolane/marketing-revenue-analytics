package dto

type CreateCampaignRequest struct {
	Name        string  `json:"name"        binding:"required,min=2,max=255"`
	Description string  `json:"description"`
	Channel     string  `json:"channel"     binding:"required"`
	Budget      float64 `json:"budget"      binding:"required,gt=0"`
	IsPublic    bool    `json:"is_public"`
	StartsAt    *string `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
}

type UpdateCampaignRequest struct {
	Name        string  `json:"name"        binding:"required,min=2,max=255"`
	Description *string `json:"description"`
	Channel     string  `json:"channel"     binding:"required"`
	Budget      float64 `json:"budget"      binding:"required,gt=0"`
	IsPublic    bool    `json:"is_public"`
	StartsAt    *string `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft active paused completed archived"`
}

type ListCampaignsQuery struct {
	Status   string `form:"status"`
	Channel  string `form:"channel"`
	From     string `form:"from"`
	To       string `form:"to"`
	IsPublic *bool  `form:"is_public"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}

type SearchQuery struct {
	Q     string `form:"q"     binding:"required,min=1"`
	Page  int    `form:"page,default=1"`
	Limit int    `form:"limit,default=20"`
}
