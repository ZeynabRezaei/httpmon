package routes

import (
	"github.com/gin-gonic/gin"
	controller "httpmon.com/first/controllers"
	middleware "httpmon.com/first/middleware"
)

//UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/signup", controller.SignUp())
	incomingRoutes.POST("/users/login", controller.Login())
	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.GET("/users/history", controller.GetHistory())
}
