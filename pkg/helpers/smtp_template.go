package helpers

import (
	"html/template"

	"bytes"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type Email struct {
	Recipient []string
	Body      interface{}
	Template  []byte
}

// func SendEmail(emailChan chan *Email) {
// 	var email = <-emailChan
// 	var buf bytes.Buffer

// 	// Convert the byte slice to a string
// 	htmlStr := string(email.Template)
// 	//  Parse the template
// 	tmpl, err := template.New("email").Parse(htmlStr)
// 	if err != nil {
// 		return
// 	}
// 	if err := tmpl.Execute(&buf, email.Body); err != nil {
// 		return
// 	}
// 	mailer := gomail.NewMessage()
// 	mailer.SetHeader("From", viper.GetString("smtp.sender_name"))
// 	mailer.SetHeader("To", email.Recipient...)
// 	mailer.SetHeader("Subject", "Test mail")
// 	mailer.SetBody("text/html", buf.String())
// 	dialer := gomail.NewDialer(
// 		viper.GetString("smtp.host"),
// 		viper.GetInt("smtp.port"),
// 		viper.GetString("smtp.email"),
// 		viper.GetString("smtp.password"),
// 	)

// 	err = dialer.DialAndSend(mailer)
// 	if err != nil {
// 		return
// 	}
// 	logrus.Info("Mail sent!")
// 	return
// }

func SendEmail(emailChan chan *Email) error {
	var err error
	// Run the sendEmail function in a separate goroutine
	go func() {
		var email = <-emailChan
		var buf bytes.Buffer

		// Convert the byte slice to a string
		htmlStr := string(email.Template)
		//  Parse the template
		tmpl, err := template.New("email").Parse(htmlStr)
		if err != nil {
			logrus.Error(err)
			return
		}
		if err := tmpl.Execute(&buf, email.Body); err != nil {
			logrus.Error(err)
			return
		}
		mailer := gomail.NewMessage()
		mailer.SetHeader("From", viper.GetString("smtp.sender_name"))
		mailer.SetHeader("To", email.Recipient...)
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
			logrus.Error(err)
			return
		}
		logrus.Info("Mail sent!")
	}()
	return err
}
