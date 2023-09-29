package globals

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/wonderivan/logger"
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
	MailServerHost string
	MailServerPort int
	MailFrom       string
	MailSender     string
	MailPassword   string
)

func InitMail() {
	MailServerHost = Configures.GetString("email.host")
	MailServerPort = Configures.GetInt("email.port")
	MailFrom = fmt.Sprintf("TODO APP <%s>", Configures.GetString("email.account"))
	MailSender = Configures.GetString("email.account")
	MailPassword = Configures.GetString("email.password")
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
			Configures.GetString("sql.table"),
		),
	)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     Configures.GetString("redis.address"),
		Password: Configures.GetString("redis.password"),
		DB:       Configures.GetInt("redis.database"),
	})
}
