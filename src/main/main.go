package main

import (
	"TODOList/src/Todo"
	"TODOList/src/handler"
	"log"
)
import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/sessions"
import "github.com/gin-contrib/sessions/cookie"

func initLog() {
	log.SetPrefix("[SERVER]")
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

func main() {
	initLog()

	r := gin.Default()
	store := cookie.NewStore([]byte("adecvsefslkhj"))
	r.Use(sessions.Sessions("UserSession", store))
	r.Use(handler.JwtVerify)

	manager := Todo.Manager{}
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

	r.Run("localhost:8080")
}
