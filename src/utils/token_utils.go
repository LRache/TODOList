package utils

import (
	"TODOList/src/globals"
	"TODOList/src/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"time"
)

// Token String in Redis

func UpdateUserTokenCodeString(userId int64) string {
	code := GenerateRandomTokenCode()
	globals.RedisClient.HSet("UserTokenCode", int64ToStr(userId), code)
	logger.Trace("(UpdateUserTokenCodeString)Update user token code: userid = %v, cod = \"%v\"", userId, code)
	return code
}

func GetUserTokenCodeString(userId int64) string {
	code := globals.RedisClient.HGet("UserTokenCode", int64ToStr(userId)).Val()
	return code
}

func CheckUserTokenCodeString(userId int64, tokenCodeString string) bool {
	currentCode := globals.RedisClient.HGet("UserTokenCode", int64ToStr(userId)).Val()
	return currentCode == tokenCodeString
}

// GENERATE TOKEN

// GenerateTokenFromClaims Generate token string from user claims.
func GenerateTokenFromClaims(claims *model.UserClaimsModel) string {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(globals.TokenSecret)
	if err != nil {
		logger.Warn("GenerateTokenFromClaims: Error when signed string: \"%v\"", err.Error())
		return ""
	}
	return tokenString
}

// GenerateUserTokenBasic Generate user token id if not exists, and return token string.
func GenerateUserTokenBasic(userid int64, expiresAt int64) string {
	code := GetUserTokenCodeString(userid)
	var claims model.UserClaimsModel
	claims.UserId = userid
	claims.ExpiresAt = expiresAt
	if code == "" {
		logger.Trace("(GenerateUserTokenBasic)Token code not exists: userid = \"%v\"", code)
		claims.UserTokenCode = UpdateUserTokenCodeString(userid)
	} else {
		logger.Trace("(GenerateUserTokenBasic)Token code exists: userid = \"%v\", code = \"%v\"", userid, code)
		claims.UserTokenCode = code
	}
	return GenerateTokenFromClaims(&claims)
}

func GenerateUserToken(userid int64) string {
	return GenerateUserTokenBasic(userid, time.Now().Add(globals.UserTokenValidity).Unix())
}

func GenerateUserRefreshToken(userid int64) string {
	return GenerateUserTokenBasic(userid, time.Now().Add(globals.UserRefreshTokenValidity).Unix())
}

func GenerateEmptyUserToken() string {
	return GenerateTokenFromClaims(&model.UserClaimsModel{UserId: -1})
}

// GET TOKEN and ID

func GetUserIdFromContext(ctx *gin.Context) int64 {
	userClaimsInterface, ok := ctx.Get("userClaims")
	if !ok {
		logger.Warn("(GetUserIdFromToken)Token error.")
		return -1
	}
	userClaims := userClaimsInterface.(*model.UserClaimsModel)
	if userClaims.UserId == -1 {
		logger.Trace("(GetUserIdFromToken)Empty user.")
		return -1
	}
	if userClaims.ExpiresAt < time.Now().Unix() {
		logger.Trace("(GetUserIdFromToken)Token expired, expired at %v", userClaims.ExpiresAt)
		return -1
	} else {
		if CheckUserTokenCodeString(userClaims.UserId, userClaims.UserTokenCode) {
			return userClaims.UserId
		} else {
			return -1
		}
	}
}

func GetUserIdFromContextIgnoreExpiration(ctx *gin.Context) (int64, bool) {
	userClaimsInterface, ok := ctx.Get("userClaims")
	if !ok {
		logger.Warn(
			"(GetUserIdFromTokenIgnoreExpiration)"+
				"Context has not token or token is invalid, tokenString = \"%v\"", ctx.GetString("tokenString"))
		return -1, true
	}
	userClaims := userClaimsInterface.(*model.UserClaimsModel)
	if userClaims.UserId == -1 {
		return -1, false
	}
	return userClaims.UserId, userClaims.ExpiresAt < time.Now().Unix()
}

func GetUserTokenFromContext(ctx *gin.Context) string {
	s, e := ctx.Get("tokenString")
	if e {
		return s.(string)
	}
	u, e := ctx.Get("userClaims")
	if !e {
		return GenerateEmptyUserToken()
	} else {
		userClaims := u.(*model.UserClaimsModel)
		return GenerateTokenFromClaims(userClaims)
	}
}

func GetUserRefreshTokenStringFromContext(ctx *gin.Context) string {
	s, _ := ctx.Get("refreshTokenString")
	return s.(string)
}
