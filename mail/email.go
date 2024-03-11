package mail

import (
	"fmt"
	"net/smtp"
	"os"
)

type Email struct {
	auth smtp.Auth
}

func NewEmailConnection() *Email {
	email_username := os.Getenv("EMAIL")
	email_password := os.Getenv("EMAILPASSWORD")
	auth := smtp.PlainAuth("", email_username, email_password, "smtp.gmail.com")
	return &Email{
		auth: auth,
	}
}

func (email *Email) SendMessage(from, subject, body string, to []string) error {
	msg := "Subject: " + subject + "\n" + body
	return smtp.SendMail("smtp.gmail.com:587", email.auth, from, to, []byte(msg))
}

func (email *Email) EmailConfirmationMessage(to string, appeal, link string) error {
	msg := "Subject: [Libraria] Confirm registration\n"
	body := fmt.Sprintf("Dear %s!\n\nTo complete registration please follow the link:\n\n %s", appeal, link)
	msg += body
	return smtp.SendMail("smtp.gmail.com:587", email.auth, "Libraria", []string{to}, []byte(msg))
}

func (email *Email) PasswordResetMessage(to string, appeal, link string) error {
	msg := "Subject: [Libraria] Password reset\n"
	body := fmt.Sprintf("Dear %s!\n\nAs you have requested for reset password instructions, here they are, please follow the URL:\n\n%s", appeal, link)
	msg += body
	return smtp.SendMail("smtp.gmail.com:587", email.auth, "Libraria", []string{to}, []byte(msg))
}
