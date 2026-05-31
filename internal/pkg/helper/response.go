package helper

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/pkg/validator"
)

func Success(ctx *gin.Context, status int, message string, data any) {
	requestID, _ := ctx.Get("X-Request-ID")
	requestIDStr, _ := requestID.(string)

	if requestIDStr != "" {
		ctx.Header("X-Request-ID", requestIDStr)
	}

	// logger.Log.WithFields(logrus.Fields{
	// 	"status":     status,
	// 	"path":       ctx.Request.URL.Path,
	// 	"method":     ctx.Request.Method,
	// 	"request_id": requestIDStr,
	// }).Info(message)

	ctx.JSON(status, dto.APIResponse[any]{
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

		// if appErr, ok := util.IsAppError(err); ok {
		// 	if status == http.StatusInternalServerError || status == 0 {
		// 		statusCode = appErr.Code
		// 	}
		// 	message = appErr.Message
		// }
	}

	requestID, _ := ctx.Get("X-Request-ID")
	requestIDStr, _ := requestID.(string)

	if requestIDStr != "" {
		ctx.Header("X-Request-ID", requestIDStr)
	}

	// logFields := logrus.Fields{
	// 	"status":     statusCode,
	// 	"path":       ctx.Request.URL.Path,
	// 	"method":     ctx.Request.Method,
	// 	"error":      errDetail,
	// 	"request_id": requestIDStr,
	// }

	// if statusCode >= 500 {
	// 	logger.Log.WithFields(logFields).Error(message)
	// } else {
	// 	logger.Log.WithFields(logFields).Warn(message)
	// }

	ctx.JSON(statusCode, dto.APIResponse[any]{
		Success: false,
		Code:    statusCode,
		Message: message,
		Data:    detailData,
	})
}
