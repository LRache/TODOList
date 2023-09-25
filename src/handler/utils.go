package handler

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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
		ctx.Set("userid", -1)
	} else {
		ctx.Set("userid", ParseToken(t).Id)
	}
}

func ParseToken(tokenString string) *UserClaims {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		log.Printf("ParseToken: %v\n", err.Error())
		return &UserClaims{Id: -1}
	}
	claims, ok := token.Claims.(*UserClaims)
	if ok {
		return claims
	} else {
		log.Printf("ParseToken: Invalid token\n")
		return &UserClaims{Id: -1}
	}
}

func GetUserIdFromToken(ctx *gin.Context) int {
	userIdInterface, ok := ctx.Get("userid")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Token error"})
		log.Printf("Manager.RequestGetAllItem: Token error.")
		return -1
	}
	return userIdInterface.(int)
}
