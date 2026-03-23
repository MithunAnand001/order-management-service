package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"order-management-service/internal/config"
	"order-management-service/internal/utils"

	"go.uber.org/zap"
)

type CommunicationService interface {
	SendOrderConfirmationEmail(ctx context.Context, userName, to, orderUUID string, amount float64) error
}

type commSer struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewCommunicationService(cfg *config.Config, logger *zap.Logger) CommunicationService {
	return &commSer{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *commSer) SendOrderConfirmationEmail(ctx context.Context, userName, to, orderUUID string, amount float64) error {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start CommunicationService.SendOrderConfirmationEmail", zap.String("request_id", reqID), zap.String("to", to))

	subject := fmt.Sprintf("Order Confirmation: %s", orderUUID)
	body := fmt.Sprintf("Hi %s,\n\nThank you for your order!\n\nOrder ID: %s\nTotal Amount: $%.2f\n\nYour order has been placed successfully.", userName, orderUUID, amount)

	header := make(map[string]string)
	header["From"] = s.cfg.FromEmail
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	var message strings.Builder
	for k, v := range header {
		fmt.Fprintf(&message, "%s: %s\r\n", k, v)
	}
	message.WriteString("\r\n" + body)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// Gmail Port 465 requires TLS from the start
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.cfg.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		s.logger.Error("TLS Dial error", zap.Error(err))
		return err
	}

	c, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		s.logger.Error("SMTP Client error", zap.Error(err))
		return err
	}

	if err = c.Auth(auth); err != nil {
		s.logger.Error("SMTP Auth error", zap.Error(err))
		return err
	}

	if err = c.Mail(s.cfg.FromEmail); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message.String()))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()

	s.logger.Info("Email sent successfully to Gmail", zap.String("request_id", reqID))
	return nil
}
