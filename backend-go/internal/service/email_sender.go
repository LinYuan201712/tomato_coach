package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

type EmailSender interface {
	SendVerificationCode(ctx context.Context, toEmail, code string, ttl time.Duration) error
}

type SMTPEmailConfig struct {
	Enabled     bool
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	UseTLS      bool
}

type smtpEmailSender struct {
	cfg    SMTPEmailConfig
	logger *logger.Logger
}

func NewSMTPEmailSender(cfg SMTPEmailConfig, logger *logger.Logger) EmailSender {
	return &smtpEmailSender{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *smtpEmailSender) SendVerificationCode(ctx context.Context, toEmail, code string, ttl time.Duration) error {
	if !s.cfg.Enabled || strings.TrimSpace(s.cfg.Host) == "" {
		return errors.New(errors.CodeInternalError, "邮件服务未配置")
	}
	if strings.TrimSpace(s.cfg.Username) == "" || strings.TrimSpace(s.cfg.Password) == "" {
		return errors.New(errors.CodeInternalError, "邮件账号未配置")
	}
	if strings.TrimSpace(s.cfg.FromAddress) == "" {
		return errors.New(errors.CodeInternalError, "发件人邮箱未配置")
	}

	from := mail.Address{Name: s.cfg.FromName, Address: s.cfg.FromAddress}
	to := mail.Address{Address: toEmail}
	subject := "番茄学习助手邮箱验证码"
	body := fmt.Sprintf("您的验证码是：%s\n\n验证码将在 %d 分钟后过期。若非本人操作，请忽略本邮件。", code, int(ttl.Minutes()))
	msg := buildEmailMessage(from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	if err := s.send(ctx, addr, auth, from.Address, []string{to.Address}, msg); err != nil {
		s.logger.Error("发送邮箱验证码失败", zap.String("email", toEmail), zap.Error(err))
		return errors.New(errors.CodeInternalError, "发送验证码失败，请稍后再试")
	}

	return nil
}

func (s *smtpEmailSender) send(ctx context.Context, addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	if s.cfg.UseTLS || s.cfg.Port == 465 {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return err
		}
		return sendWithClient(tlsConn, host, auth, from, to, msg)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func sendWithClient(conn net.Conn, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildEmailMessage(from, to mail.Address, subject, body string) []byte {
	var b bytes.Buffer
	b.WriteString("From: " + from.String() + "\r\n")
	b.WriteString("To: " + to.String() + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")
	return b.Bytes()
}
