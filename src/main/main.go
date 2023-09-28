package main

import (
	"TODOList/src/ServerManager"
	"TODOList/src/globals"
	"TODOList/src/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
)

func initLog() {
	log.SetPrefix("[SERVER]")
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {
	initLog()
	globals.InitConfigures("./configures.yml")

	r := gin.Default()
	store := cookie.NewStore([]byte("adecvsefslkhj"))
	r.Use(sessions.Sessions("UserSession", store))
	r.Use(handler.JwtVerify)

	manager := ServerManager.Manager{}
	manager.Init()
	defer manager.End()

	// items
	r.GET("/todo/item/:id", manager.RequestGetItemById)
	r.GET("/todo/item", manager.RequestGetItems)
	r.PUT("/todo/item", manager.RequestAddItem)
	r.POST("/todo/item", manager.RequestUpdateItem)
	r.DELETE("/todo/item/:id", manager.RequestDeleteItemById)

	r.GET("/todo/user", manager.RequestGetCurrentUser)
	r.PUT("/todo/user", manager.RequestRegisterUser)
	r.POST("/todo/user", manager.RequestLogin)
	r.DELETE("/todo/user", manager.RequestDeleteUser)
	r.POST("/todo/user/token", manager.RequestRefreshToken)

	r.Run(globals.Configures.GetString("server.host") +
		":" + globals.Configures.GetString("server.port"))
}
