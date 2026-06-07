package helper

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type SwaggerDocResponse struct {
	Success bool   `json:"success" example:"true"`
	Code    int    `json:"code"    example:"200"`
	Message string `json:"message" example:"success"`
	Data    any    `json:"data"`
}

func Success(ctx *gin.Context, status int, message string, data any) {
	requestID, _ := ctx.Get("X-Request-ID")
	requestIDStr, _ := requestID.(string)

	if requestIDStr != "" {
		ctx.Header("X-Request-ID", requestIDStr)
	}

	ctx.JSON(status, APIResponse[any]{
		Success: true,
		Code:    status,
		Message: message,
		Data:    data,
	})
}

func Error(ctx *gin.Context, status int, message string, err error) {
	statusCode := status
	var detailData any = nil

	if err != nil {
		translatedErrs := validator.TranslateValidationError(err)
		if len(translatedErrs) > 0 {
			detailData = translatedErrs
		}
	}

	requestID, _ := ctx.Get("X-Request-ID")
	requestIDStr, _ := requestID.(string)

	if requestIDStr != "" {
		ctx.Header("X-Request-ID", requestIDStr)
	}

	ctx.JSON(statusCode, APIResponse[any]{
		Success: false,
		Code:    statusCode,
		Message: message,
		Data:    detailData,
	})
}
