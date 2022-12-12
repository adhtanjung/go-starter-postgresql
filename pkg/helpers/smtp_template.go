package helpers

import (
	"fmt"
	"html/template"

	"bytes"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func SendEmail(emailTemplate []byte, token string, emails ...string) (err error) {
	var buf bytes.Buffer

	// Convert the byte slice to a string
	htmlStr := string(emailTemplate)
	//  Parse the template
	tmpl, err := template.New("email").Parse(htmlStr)
	if err != nil {
		return err
	}
	// Define the data that will be used to fill the template
	data := struct {
		ResetPasswordLink string
	}{
		ResetPasswordLink: fmt.Sprintf("https://example.com/reset-password?token=%s", token),
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
