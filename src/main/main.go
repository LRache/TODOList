package main

import (
	"TODOList/src/globals"
	"TODOList/src/handler"
	"TODOList/src/server"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/wonderivan/logger"
)

func main() {
	globals.InitConfigures("./configures.yml")
	globals.InitLogger()
	globals.InitMail()
	globals.InitDatabase()
	defer globals.End()

	router := gin.Default()
	store := cookie.NewStore([]byte("adecvsefslkhj"))
	router.Use(sessions.Sessions("UserSession", store))
	router.Use(handler.JwtVerify)

	// item
	itemGroup := router.Group("/todo/item")
	itemGroup.GET("/:id", server.RequestGetItemById)
	itemGroup.GET("", server.RequestGetItems)
	itemGroup.PUT("", server.RequestAddItem)
	itemGroup.POST("", server.RequestUpdateItem)
	itemGroup.DELETE("/:id", server.RequestDeleteItemById)
	itemGroup.GET("/cron", server.RequestSetItemCron)

	// user
	userGroup := router.Group("/todo/user")
	userGroup.GET("", server.RequestGetCurrentUser)
	userGroup.PUT("", server.RequestRegisterUser)
	userGroup.POST("", server.RequestLogin)
	userGroup.DELETE("", server.RequestDeleteUser)
	userGroup.POST("/token", server.RequestRefreshToken)

	userGroup.GET("/mail", server.RequestSendVerifyMail)
	userGroup.POST("/mail", server.RequestGetMailVerify)
	userGroup.POST("/reset", server.RequestResetUser)

	// Scan and remove expired verify code every 10 minutes.
	c := cron.New()
	_, err := c.AddFunc("@every 10m", server.RemoveExpiredVerifyCode)
	if err != nil {
		logger.Error("Error when add remove expired verify code function to cron.")
	} else {
		c.Start()
		logger.Trace("Cron in main function start successfully.")
	}

	err = router.Run(globals.Configures.GetString("server.host") +
		":" + globals.Configures.GetString("server.port"))
	if err != nil {
		logger.Emer("Run server error.")
	}
}
