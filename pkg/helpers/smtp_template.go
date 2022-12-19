package helpers

import (
	"html/template"

	"bytes"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func SendEmail(emailTemplate []byte, data interface{}, emails ...string) (err error) {
	var buf bytes.Buffer

	// Convert the byte slice to a string
	htmlStr := string(emailTemplate)
	//  Parse the template
	tmpl, err := template.New("email").Parse(htmlStr)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", viper.GetString("smtp.sender_name"))
	mailer.SetHeader("To", emails...)
	mailer.SetHeader("Subject", "Test mail")
	mailer.SetBody("text/html", buf.String())
	dialer := gomail.NewDialer(
		viper.GetString("smtp.host"),
		viper.GetInt("smtp.port"),
		viper.GetString("smtp.email"),
		viper.GetString("smtp.password"),
	)

	err = dialer.DialAndSend(mailer)
	if err != nil {
		return err
	}
	logrus.Info("Mail sent!")
	return
}
