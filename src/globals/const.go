package globals

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	UserTokenValidity        = time.Hour * 24
	UserRefreshTokenValidity = time.Hour * 144
)

const (
	StatusDatabaseCommandOK      = 0
	StatusDatabaseCommandError   = 1
	StatusDatabaseSelectNotFound = 2
)

type ReturnJson struct {
	Code int
	Json gin.H
}

func returnJson(code int, message string) ReturnJson {
	return ReturnJson{code, gin.H{"code": code, "message": message}}
}

var (
	ReturnJsonUserNotLogin        = returnJson(http.StatusUnauthorized, "User not login.")
	ReturnJsonInternalServerError = returnJson(http.StatusInternalServerError, "Internal server error when insert item.")
	ReturnJsonItemNotFound        = returnJson(http.StatusBadRequest, "Item not found.")
	ReturnJsonParamError          = returnJson(http.StatusBadRequest, "Param error.")
	ReturnJsonQueryError          = returnJson(http.StatusBadRequest, "Query error.")
	ReturnJsonBodyJsonError       = returnJson(http.StatusBadRequest, "Body json error.")
	ReturnJsonSuccess             = returnJson(http.StatusOK, "Success")
)
