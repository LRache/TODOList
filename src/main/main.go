package main

import (
	"TODOList/src/ServerManager"
	"TODOList/src/globals"
	"TODOList/src/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
)

func main() {
	globals.InitConfigures("./configures.yml")
	globals.InitLogger()
	globals.InitMail()

	router := gin.Default()
	store := cookie.NewStore([]byte("adecvsefslkhj"))
	router.Use(sessions.Sessions("UserSession", store))
	router.Use(handler.JwtVerify)

	manager := ServerManager.Manager{}
	manager.Init()
	defer manager.End()

	// item
	itemGroup := router.Group("/todo/item")
	itemGroup.GET("/:id", manager.RequestGetItemById)
	itemGroup.GET("", manager.RequestGetItems)
	itemGroup.PUT("", manager.RequestAddItem)
	itemGroup.POST("", manager.RequestUpdateItem)
	itemGroup.DELETE("/:id", manager.RequestDeleteItemById)

	// user
	userGroup := router.Group("/todo/user")
	userGroup.GET("", manager.RequestGetCurrentUser)
	userGroup.PUT("", manager.RequestRegisterUser)
	userGroup.POST("", manager.RequestLogin)
	userGroup.DELETE("", manager.RequestDeleteUser)
	userGroup.POST("/token", manager.RequestRefreshToken)

	userGroup.GET("/mail", manager.RequestSendVerifyMail)
	userGroup.POST("/mail", manager.RequestGetMailVerify)

	err := router.Run(globals.Configures.GetString("server.host") +
		":" + globals.Configures.GetString("server.port"))
	if err != nil {
		logger.Emer("Run server error.")
	}
}
