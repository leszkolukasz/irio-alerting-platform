package email

import (
	"alerting-platform/common/config"
	"context"
	"fmt"
	"log"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func Init(ctx context.Context) (*Mailer, error) {
	cfg := config.GetConfig()

	host, portStr, user, password, fromEmail := cfg.SmtpHost, cfg.SmtpPort, cfg.SmtpUser, cfg.SmtpPass, cfg.EmailFrom

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp port: %w", err)
	}

	d := gomail.NewDialer(host, port, user, password)

	return &Mailer{
		dialer: d,
		from:   fromEmail,
	}, nil
}

func (m *Mailer) SendNotification(toEmail string, incidentID string, serviceID uint64) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", fmt.Sprintf("[ALERT] New Incident: %s", incidentID))

	body := fmt.Sprintf(`
		<h2>You've got a new incident!</h2>
		<p><strong>ID:</strong> %s</p>
		<p><strong>Service:</strong> %s</p>
		<p>Follow this link to view the incident details:</p>
	`, incidentID, strconv.FormatUint(serviceID, 10))

	msg.SetBody("text/html", body)

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email via gomail: %w", err)
	}

	log.Printf("[INFO] Notification email sent to %s for incident %s", toEmail, incidentID)

	return nil
}
