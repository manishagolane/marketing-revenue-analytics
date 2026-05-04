package dto

import "marketing-revenue-analytics/constants"

type APIResponse struct {
	Status  constants.Status `json:"status"`
	Message string           `json:"message"`
	Data    interface{}      `json:"data,omitempty"`
}
