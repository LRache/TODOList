package globals

import "github.com/wonderivan/logger"

func InitLogger() {
	err := logger.SetLogger(`{"Console": {"level": "TRAC", "color": true}}`)
	if err != nil {
		return
	}
}
