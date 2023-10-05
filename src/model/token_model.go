package model

import "github.com/dgrijalva/jwt-go"

type UserClaimsModel struct {
	UserTokenCode string `json:"userTokenCode"`
	UserId        int64  `json:"userId"`
	jwt.StandardClaims
}
