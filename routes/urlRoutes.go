package routes

import (
	"github.com/gin-gonic/gin"
	controller "httpmon.com/first/controllers"
	middleware "httpmon.com/first/middleware"
)

//urlRoutes function
func UrlRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.POST("/url/create", controller.CreateUrl())
	incomingRoutes.POST("/url/delete", controller.DeleteUrl())
	incomingRoutes.GET("/url/get", controller.GetUrl())
}
