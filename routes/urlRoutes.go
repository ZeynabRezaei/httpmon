package routes

import (
	"github.com/gin-gonic/gin"
	controller "httpmon.com/first/controllers"
	middleware "httpmon.com/first/middleware"
)

//urlRoutes function
func UrlRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.POST("/Url/create", controller.CreateUrl())
	incomingRoutes.GET("/Url/get", controller.GetUrl())
}
