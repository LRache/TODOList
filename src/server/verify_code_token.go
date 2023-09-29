package server

import (
	"TODOList/src/globals"
	"github.com/dgrijalva/jwt-go"
	"github.com/wonderivan/logger"
)

type MailVerifyCodeClaims struct {
	MailAddr string `json:"mail"`
	jwt.StandardClaims
}

func GenerateMailVerifyCodeToken(claims *MailVerifyCodeClaims) (string, bool) {
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(globals.TokenSecret)
	if err != nil {
		logger.Warn("(GenerateMailVerifyCodeToken)Error when signed string: %v", err.Error())
		return "", false
	}
	return t, true
}

func GetMailAddrFromToken(tokenString string) (string, bool) {
	t, err := jwt.ParseWithClaims(tokenString, &MailVerifyCodeClaims{}, func(token *jwt.Token) (interface{}, error) {
		return globals.TokenSecret, nil
	})
	if err != nil {
		logger.Warn("(GetMailAddrFromToken)Error when parse token: %v", err.Error())
		return "", false
	}
	claims, ok := t.Claims.(*MailVerifyCodeClaims)
	if ok {
		logger.Trace("(GetMailAddrFromToken)Parse token successfully.")
		return claims.MailAddr, true
	} else {
		logger.Warn("(GetMailAddrFromToken)Error when parse token.")
		return "", false
	}
}
