package handler

import (
	"net/http"

	"10.1.20.130/dropping/user-service/docs"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterUserRoutes(r *gin.Engine, uh UserHandler) *gin.Engine {
	docs.SwaggerInfo.BasePath = "/api/v1/user"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/healthy", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Healthy")
	})
	{
		r.PATCH("", uh.UpdateUser)
		r.DELETE("", uh.DeleteUser)
		r.PATCH("/email", uh.ChangeEmail)
		r.PATCH("/password", uh.ChangePassword)
		r.GET("/me", uh.GetProfile)
	}
	return r
}
