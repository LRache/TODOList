package utils

import (
	"TODOList/src/globals"
	"crypto/md5"
	"fmt"
	"strconv"
)

func IsValidUsername(name string) bool {
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

func StringToMd5(password string) string {
	data := []byte(password)
	pwd := md5.Sum(data)
	return fmt.Sprintf("%x", pwd)
}

func GenerateVerifyCode() string {
	var s string
	for i := 0; i < 6; i++ {
		s += strconv.Itoa(globals.Rand.Intn(10))
	}
	return s
}
