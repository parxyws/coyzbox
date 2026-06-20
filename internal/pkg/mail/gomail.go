package mail

import (
	"bytes"
	"html/template"
	"path"

	"github.com/parxyws/cozybox/internal/config"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
	config *config.Config
}

func NewGoMailDialer(cfg *config.Config) *gomail.Dialer {
	return gomail.NewDialer(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password)
}

func NewMailer(dialer *gomail.Dialer, cfg *config.Config) *Mailer {
	return &Mailer{dialer: dialer, config: cfg}
}

func (m *Mailer) SendOTP(to string, data OTPData) error {
	return m.sendTemplate(to, "Verify Your Email", path.Join("internal", "pkg", "mail", "template", "otp.html"), data)
}

func (m *Mailer) SendResetPassword(to string, data ResetPasswordData) error {
	return m.sendTemplate(to, "Reset Password", path.Join("internal", "pkg", "mail", "template", "reset-password.html"), data)
}

func (m *Mailer) SendWelcomeEmail(to, subject string, data OTPData) error {
	return m.sendTemplate(to, subject, path.Join("internal", "pkg", "mail", "template", "welcome.html"), data)
}

func (m *Mailer) sendTemplate(to, subject, filePath string, data any) error {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.config.Mail.User)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body.String())
	return m.dialer.DialAndSend(message)
}
