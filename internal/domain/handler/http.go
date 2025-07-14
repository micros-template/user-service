package handler

import "github.com/gin-gonic/gin"

func RegisterUserRoutes(r *gin.Engine, uh UserHandler) *gin.Engine {
	{
		r.PATCH("", uh.UpdateUser)
		r.PATCH("/email", uh.ChangeEmail)
		r.PATCH("/password", uh.ChangePassword)
		r.GET("/me", uh.GetProfile)
	}
	return r
}
