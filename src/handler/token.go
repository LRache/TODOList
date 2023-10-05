package handler

import (
	"TODOList/src/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
)

func JwtVerify(ctx *gin.Context) {
	t := ctx.GetHeader("token")
	if t == "" {
		logger.Trace("(JwtVerify)Headers have no token.")
		ctx.Set("userClaims", &model.UserClaimsModel{UserId: -1})
	} else {
		ctx.Set("userTokenString", t)
		ctx.Set("userClaims", ParseToken(t))
	}
	ctx.Set("tokenString", t)
	refreshToken := ctx.GetHeader("refreshToken")
	ctx.Set("refreshTokenString", refreshToken)
}

// ParseToken tokenString -> UserClaimsModel
func ParseToken(tokenString string) *model.UserClaimsModel {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &model.UserClaimsModel{})
	if err != nil {
		logger.Warn("(ParseToken)Error when parse token, invalid token: err = \"%v\", tokenString = \"%v\"", err.Error())
		return &model.UserClaimsModel{UserId: -1}
	}
	claims, ok := token.Claims.(*model.UserClaimsModel)
	if ok {
		logger.Trace("(ParseToken)Parse token successfully.")
		return claims
	} else {
		logger.Warn("(ParseToken)Invalid token, tokenString = \"%v\"", tokenString)
		return &model.UserClaimsModel{UserId: -1}
	}
}
