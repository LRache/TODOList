package Todo

import (
	"TODOList/src/handler"
	"crypto/md5"
	"fmt"
)

func isValidUsername(name string) bool {
	if len(name) == 0 {
		return false
	}

	r := []rune(name)
	count := 0
	for _, char := range r {
		if char == '\n' || char == '\t' || char == '\r' {
			return false
		} else if char == ' ' {
			count++
		}
	}
	return count != len(r)
}

func toMd5(password string) string {
	data := []byte(password)
	pwd := md5.Sum(data)
	return fmt.Sprintf("%x", pwd)
}

func GenerateUserToken(userid int) string {
	return handler.GenerateToken(&handler.UserClaims{Id: userid})
}

func GenerateNoUserToken() string {
	return handler.GenerateToken(&handler.UserClaims{Id: -1})
}
