package handler

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type UserClaims struct {
	Id int `json:"userid"`
	jwt.StandardClaims
}

var (
	secret = []byte("zjdxfszx20200635")
)

func GenerateToken(claims *UserClaims) string {
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		log.Printf("GenerateToken: Error when signed string: %v\n", err.Error())
		return ""
	}
	return t
}

func JwtVerify(ctx *gin.Context) {
	t := ctx.GetHeader("token")
	if t == "" {
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
		log.Printf("ParseToken: %v\n", err.Error())
		return &UserClaims{Id: -1}
	}
	claims, ok := token.Claims.(*UserClaims)
	if ok {
		return claims
	} else {
		log.Printf("ParseToken: Invalid token.\n")
		return &UserClaims{Id: -1}
	}
}

func GetUserIdFromToken(ctx *gin.Context) int {
	userClaimsInterface, ok := ctx.Get("userClaims")
	if !ok {
		log.Printf("Manager.RequestGetAllItem: Token error.")
		return -1
	}
	userClaims := userClaimsInterface.(*UserClaims)
	if userClaims.Id == -1 {
		return -1
	}
	if userClaims.ExpiresAt < time.Now().Unix() {
		return -1
	} else {
		return userClaims.Id
	}
}

func GetUserIdFromTokenIgnoreExpiration(ctx *gin.Context) (int, bool) {
	userClaimsInterface, ok := ctx.Get("userClaims")
	if !ok {
		log.Printf("Manager.RequestGetAllItem: Token error.")
		return -1, true
	}
	userClaims := userClaimsInterface.(*UserClaims)
	if userClaims.Id == -1 {
		return -1, false
	}
	return userClaims.Id, userClaims.ExpiresAt < time.Now().Unix()
}
