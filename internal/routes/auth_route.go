package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/handlers"
)

func AuthRoute(route *gin.RouterGroup, handler *handlers.AuthHandler) {
	auth := route.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/verify-email", handler.VerifyEmail)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh-token", handler.RefreshToken)
		auth.POST("/forgot-password", handler.ForgotPassword)
		auth.POST("/reset-password", handler.ResetPassword)
	}
}

// AuthProtectedRoute registers auth routes that require authentication.
func AuthProtectedRoute(route *gin.RouterGroup, handler *handlers.AuthHandler) {
	auth := route.Group("/auth")
	{
		auth.POST("/logout", handler.Logout)
	}
}
