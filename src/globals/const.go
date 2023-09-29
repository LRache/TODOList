package globals

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

const (
	UserTokenValidity        = time.Hour * 24
	UserRefreshTokenValidity = time.Hour * 144
)

const (
	StatusInternalServerError = -1
	StatusOK                  = 0
)

const (
	StatusDatabaseCommandOK      = 0
	StatusDatabaseCommandError   = 1
	StatusDatabaseSelectNotFound = 2
)

const (
	StatusNoVerifyCode        = 1
	StatusIncorrectVerifyCode = 2
)

type ReturnJson struct {
	Code int
	Json gin.H
}

func NewReturnJson(code int, message string) ReturnJson {
	return ReturnJson{code, gin.H{"code": code, "message": message}}
}

var (
	ReturnJsonUserNotLogin        = NewReturnJson(http.StatusUnauthorized, "User not login.")
	ReturnJsonInternalServerError = NewReturnJson(http.StatusInternalServerError, "Internal server error.")
	ReturnJsonItemNotFound        = NewReturnJson(http.StatusNotFound, "Item not found.")
	ReturnJsonParamError          = NewReturnJson(http.StatusBadRequest, "Param error.")
	ReturnJsonQueryError          = NewReturnJson(http.StatusBadRequest, "Query error.")
	ReturnJsonBodyJsonError       = NewReturnJson(http.StatusBadRequest, "Body json error.")
	ReturnJsonSuccess             = NewReturnJson(http.StatusOK, "Success")
)

var Rand = rand.New(rand.NewSource(time.Now().Unix()))

var TokenSecret = []byte("zjdxfszx20200635")
