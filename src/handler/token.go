package handler

import (
	"TODOList/src/globals"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"time"
)

type UserClaims struct {
	Id int64 `json:"userid"`
	jwt.StandardClaims
}

func GenerateToken(claims *UserClaims) string {
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(globals.TokenSecret)
	if err != nil {
		logger.Warn("GenerateToken: Error when signed string: %v", err.Error())
		return ""
	}
	return t
}

func JwtVerify(ctx *gin.Context) {
	t := ctx.GetHeader("token")
	if t == "" {
		logger.Trace("(JwtVerify)Headers have no token.")
		ctx.Set("userClaims", &UserClaims{Id: -1})
	} else {
		ctx.Set("userClaims", ParseToken(t))
	}
	ctx.Set("tokenString", t)
	refreshToken := ctx.GetHeader("refreshToken")
	ctx.Set("refreshTokenString", refreshToken)
}

func ParseToken(tokenString string) *UserClaims {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &UserClaims{})
	if err != nil {
		logger.Warn("(ParseToken)Error when parse token, invalid token: %v", err.Error())
		return &UserClaims{Id: -1}
	}
	claims, ok := token.Claims.(*UserClaims)
	if ok {
		logger.Trace("(ParseToken)Parse token successfully.")
		return claims
	} else {
		logger.Warn("(ParseToken)Invalid token, tokenString = %v", tokenString)
		return &UserClaims{Id: -1}
	}
}

func GetUserIdFromToken(ctx *gin.Context) int64 {
	userClaimsInterface, ok := ctx.Get("userClaims")
	if !ok {
		logger.Warn("(GetUserIdFromToken)Token error.")
		return -1
	}
	userClaims := userClaimsInterface.(*UserClaims)
	if userClaims.Id == -1 {
		logger.Trace("(GetUserIdFromToken)Empty user.")
		return -1
	}
	if userClaims.ExpiresAt < time.Now().Unix() {
		logger.Trace("(GetUserIdFromToken)Token expired, expired at %v", userClaims.ExpiresAt)
		return -1
	} else {
		return userClaims.Id
	}
}

func GetUserIdFromTokenIgnoreExpiration(ctx *gin.Context) (int64, bool) {
	userClaimsInterface, ok := ctx.Get("userClaims")
	if !ok {
		logger.Warn(
			"(GetUserIdFromTokenIgnoreExpiration)"+
				"Context has not token or token is invalid, tokenString = \"%v\"", ctx.GetString("tokenString"))
		return -1, true
	}
	userClaims := userClaimsInterface.(*UserClaims)
	if userClaims.Id == -1 {
		return -1, false
	}
	return userClaims.Id, userClaims.ExpiresAt < time.Now().Unix()
}
