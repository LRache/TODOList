package utils

import (
	"TODOList/src/globals"
	"crypto/md5"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/wonderivan/logger"
	"regexp"
	"strconv"
	"strings"
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

func IsMailFormat(s string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(s)
}

func GenerateOnceCron(c *cron.Cron, spec string, f func()) error {
	var id cron.EntryID
	id, err := c.AddFunc(spec, func() {
		f()
		c.Remove(id)
	})
	return err
}

func StringToMd5(password string) string {
	data := []byte(password)
	pwd := md5.Sum(data)
	return fmt.Sprintf("%x", pwd)
}

func GenerateRandomTokenCode() string {
	code := make([]string, 16)
	for i := 0; i < 16; i++ {
		n := globals.Rand.Intn(62)
		if n < 10 {
			code[i] = string(byte(48 + n))
		} else {
			if n < 36 {
				code[i] = string(byte(55 + n))
			} else {
				code[i] = string(byte(61 + n))
			}
		}
	}
	return strings.Join(code, "")
}

func GenerateRandomVerifyCode() string {
	var s string
	for i := 0; i < 6; i++ {
		s += strconv.Itoa(globals.Rand.Intn(10))
	}
	return s
}

func int64ToStr(v int64) string {
	return fmt.Sprintf("%v", v)
}

func GetUserCount() int64 {
	count, err := globals.RedisClient.Get("UserCount").Int64()
	if err != nil {
		logger.Alert("(GetUserCount)Error when get from redis: %v", err.Error())
		return -1
	}
	return count
}

func SetUserCount(count int64) {
	globals.RedisClient.Set("UserCount", count, 0)
}

func SetUserCountPlusOne() int64 {
	v := GetUserCount() + 1
	SetUserCount(v)
	return v
}

func GetItemCount(userId int64) int64 {
	count, err := globals.RedisClient.HGet("ItemCount", fmt.Sprintf("%v", userId)).Int64()
	if err != nil {
		logger.Alert("(GetItemCount)Error when get from redis: %v", err.Error())
		return -1
	}
	return count
}

func SetItemCount(userId int64, count int64) {
	globals.RedisClient.HSet("ItemCount", int64ToStr(userId), count)
}

func SetItemCountPlusOne(userId int64) {
	SetItemCount(userId, GetItemCount(userId)+1)
}
