package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"marketing-revenue-analytics/config"
	"marketing-revenue-analytics/constants"
	"marketing-revenue-analytics/internal/dto"
)

type RequestError struct {
	Status    constants.Status `json:"status"`
	Message   string           `json:"message"`
	ErrorCode int              `json:"errorCode"`
}

func SendSuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  constants.ApiSuccess,
		Message: message,
		Data:    data,
	})
}

func SendBadRequestError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, RequestError{
		Status:    constants.ApiFailure,
		Message:   getPrettyValidationError(err).Error(),
		ErrorCode: http.StatusBadRequest,
	})
}

func SendValidationError(c *gin.Context, err error) {
	SendBadRequestError(c, err)
}

func SendUnauthorizedError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, RequestError{
		Status:    constants.ApiFailure,
		Message:   err.Error(),
		ErrorCode: http.StatusUnauthorized,
	})
}

func SendForbiddenError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusForbidden, RequestError{
		Status:    constants.ApiFailure,
		Message:   err.Error(),
		ErrorCode: http.StatusForbidden,
	})
}

func SendNotFoundError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusNotFound, RequestError{
		Status:    constants.ApiFailure,
		Message:   err.Error(),
		ErrorCode: http.StatusNotFound,
	})
}

func SendConflictError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusConflict, RequestError{
		Status:    constants.ApiFailure,
		Message:   err.Error(),
		ErrorCode: http.StatusConflict,
	})
}

func SendApplicationError(c *gin.Context, err error) {
	msg := "unable to process request"
	if config.GetString("environment") != "production" {
		msg = err.Error()
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, RequestError{
		Status:    constants.ApiFailure,
		Message:   msg,
		ErrorCode: http.StatusInternalServerError,
	})
}

func SendInternalError(c *gin.Context, err error) {
	SendApplicationError(c, err)
}

func SendUnProcessableRequestError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, RequestError{
		Status:    constants.ApiFailure,
		Message:   err.Error(),
		ErrorCode: http.StatusUnprocessableEntity,
	})
}
