package lib

import (
	"net/smtp"
	"os"
)

func SendResetPassword(email string, resetLink string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, host)

	to := []string{email}

	message := []byte("To: " + email + "\r\n" +
		"Subject: Reset Password\r\n" +
		"\r\n" +
		"Click the link below to reset your password:\r\n" +
		"\r\n" +
		resetLink + "\r\n")

	err := smtp.SendMail(host+":"+port, auth, from, to, message)
	if err != nil {
		return err
	}
	return nil
}