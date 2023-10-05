package globals

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/wonderivan/logger"
	"gopkg.in/gomail.v2"
)

// CONFIGURES

var (
	Configures *viper.Viper
)

func InitConfigures(configureFilePath string) {
	Configures = viper.New()
	Configures.SetConfigFile(configureFilePath)
	err := Configures.ReadInConfig()
	if err != nil {
		logger.Alert("Read configures error.")
		return
	}
}

// LOGGER

func InitLogger() {
	err := logger.SetLogger(`{"Console": {"level": "TRAC", "color": true}}`)
	if err != nil {
		return
	}
}

// EMAIL
var (
	MailFrom   string
	MailSender *gomail.SendCloser
)

func InitMail() {
	host := Configures.GetString("email.host")
	port := Configures.GetInt("email.port")
	MailFrom = fmt.Sprintf("TODO APP <%s>", Configures.GetString("email.account"))
	username := Configures.GetString("email.account")
	password := Configures.GetString("email.password")
	d, err := gomail.NewDialer(host, port, username, password).Dial()
	MailSender = &d
	if err != nil {
		logger.Error("(InitMail)Error when dial: %v", err.Error())
	} else {
		logger.Trace("(InitMail)Mail sender dail successfully.")
	}
}

// DATABASE
var (
	SqlDatabase *sqlx.DB
	RedisClient *redis.Client
)

func InitDatabase() {
	SqlDatabase = sqlx.MustOpen(
		Configures.GetString("sql.driverName"),
		fmt.Sprintf(
			"%s:%s@tcp(%s)/%s",
			Configures.GetString("sql.userName"),
			Configures.GetString("sql.password"),
			Configures.GetString("sql.address"),
			Configures.GetString("sql.database"),
		),
	)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     Configures.GetString("redis.address"),
		Password: Configures.GetString("redis.password"),
		DB:       Configures.GetInt("redis.database"),
	})
}

func End() {
	var err error
	if err = SqlDatabase.Close(); err != nil {
		logger.Error("Error when close sql database: %v", err.Error())
	}
	if err = RedisClient.Close(); err != nil {
		logger.Error("Error when close redis: %v", err.Error())
	}
	if err = (*MailSender).Close(); err != nil {
		logger.Error("Error when close mail sender: %v", err.Error())
	}
}
