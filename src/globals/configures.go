package globals

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

var Configures *viper.Viper

func InitConfigures(configureFilePath string) {
	Configures = viper.New()
	Configures.SetConfigFile(configureFilePath)
	err := Configures.ReadInConfig()
	if err != nil {
		log.Panicln("Read configures error.")
		return
	}
	fmt.Println(Configures.Get("Server.port"))
}
