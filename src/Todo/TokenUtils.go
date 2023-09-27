package Todo

import (
	"TODOList/src/handler"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func GetUserTokenFromContext(ctx *gin.Context) string {
	s, e := ctx.Get("tokenString")
	if e {
		return s.(string)
	}
	u, e := ctx.Get("userClaims")
	if !e {
		return GenerateEmptyUserToken()
	} else {
		userClaims := u.(handler.UserClaims)
		return handler.GenerateToken(&userClaims)
	}
}

func GetUserRefreshTokenStringFromContext(ctx *gin.Context) string {
	s, _ := ctx.Get("refreshTokenString")
	return s.(string)
}

func GenerateUserTokenBasic(userid int64, expiresAt int64) string {
	var claims handler.UserClaims
	claims.Id = userid
	claims.ExpiresAt = expiresAt
	fmt.Printf("GenerateUserTokenBasic: Generate claims expire = %v\n", expiresAt)
	return handler.GenerateToken(&claims)
}

func GenerateUserToken(userid int64) string {
	return GenerateUserTokenBasic(userid, time.Now().Add(UserTokenValidity).Unix())
}

func GenerateUserRefreshToken(userid int64) string {
	return GenerateUserTokenBasic(userid, time.Now().Add(UserRefreshTokenValidity).Unix())
}

func GenerateEmptyUserToken() string {
	return handler.GenerateToken(&handler.UserClaims{Id: -1})
}
