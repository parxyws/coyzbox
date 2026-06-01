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

type OTPData struct {
	Name string
	OTP  string
}

type WelcomeData struct {
	Name         string
	TempPassword string
}

type ResetPasswordData struct {
	Name string
	Url  string
}

type InvitationData struct {
	TenantName string
	InviteUrl  string
}

func NewGoMailDialer(cfg *config.Config) *gomail.Dialer {
	return gomail.NewDialer(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password)
}

func NewMailer(dialer *gomail.Dialer, cfg *config.Config) *Mailer {
	return &Mailer{dialer: dialer, config: cfg}
}

func (m *Mailer) SendOTP(to string, subject string, data *OTPData) error {
	filePath := path.Join("internal", "tools", "mail", "template", "otp.html")
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

func (m *Mailer) SendWelcomeEmail(to string, subject string, data *WelcomeData) error {
	filePath := path.Join("internal", "tools", "mail", "template", "welcome.html")
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

func (m *Mailer) SendResetPassword(to string, subject string, data *ResetPasswordData) error {
	filePath := path.Join("internal", "tools", "mail", "template", "reset-password.html")
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

func (m *Mailer) SendChangeEmailRequest(to string, subject string, data *OTPData) error {
	filePath := path.Join("internal", "tools", "mail", "template", "change-email.html")
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

func (m *Mailer) SendChangeEmailConfirmation(to string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.config.Mail.User)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return m.dialer.DialAndSend(message)
}

func (m *Mailer) SendInvitation(to string, subject string, data *InvitationData) error {
	filePath := path.Join("internal", "tools", "mail", "template", "invitation.html")
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
