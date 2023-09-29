package globals

import "fmt"

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
