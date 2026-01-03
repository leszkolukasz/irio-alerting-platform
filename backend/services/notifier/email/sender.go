package email

import (
	"alerting-platform/common/config"
	"context"
	"fmt"
	"log"
	"strconv"

	"gopkg.in/gomail.v2"

	magic_link "alerting-platform/common/magic_link"
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
	cfg := config.GetConfig()

	resolveLink, err := magic_link.GenerateResolveLink(
		incidentID,
		serviceID,
		toEmail,
		[]byte(cfg.Secret),
		cfg.APIHost,
		cfg.REST_APIPort,
	)

	if err != nil {
		return fmt.Errorf("failed to generate resolve link: %w", err)
	}

	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", fmt.Sprintf("[ALERT] New Incident: %s", incidentID))

	body := fmt.Sprintf(`
        <div style="font-family: Arial, sans-serif; padding: 20px; max-width: 600px;">
            <h2 style="color: #d9534f;">You've got a new incident!</h2>
            <p><strong>ID:</strong> %s</p>
            <p><strong>Service:</strong> %s</p>
            
            <div style="margin: 25px 0;">
                <a href="%s" style="background-color: #d9534f; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold; display: inline-block;">
                    Resolve Incident
                </a>
            </div>

            <p style="margin-top: 30px; font-size: 13px; color: #555;">
                If the button above doesn't work, copy and paste the following URL into your browser:
            </p>
            
            <p style="font-size: 11px; color: #777; word-break: break-all; overflow-wrap: break-word; background-color: #f9f9f9; padding: 10px; border: 1px solid #eee;">
                %s
            </p>

            <p style="font-size: 12px; color: #999; margin-top: 20px;">Link is valid for 72 hours.</p>
        </div>
    `,
		incidentID,
		strconv.FormatUint(serviceID, 10),
		resolveLink,
		resolveLink,
	)

	msg.SetBody("text/html", body)

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email via gomail: %w", err)
	}

	log.Printf("[INFO] Notification email sent to %s for incident %s", toEmail, incidentID)

	return nil
}
