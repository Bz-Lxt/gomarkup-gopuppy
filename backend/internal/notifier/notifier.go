package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"gopuppy/internal/clock"
	"gopuppy/internal/config"
	"gopuppy/internal/domain"
	"gopuppy/internal/reminder"
	"gopuppy/internal/repo"
)

type Message struct {
	Channel domain.Channel
	ToEmail string
	Title   string
	Body    string
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Fanout struct {
	Email   Sender
	Wecom   Sender
	Webhook Sender
}

func New(cfg *config.Config, media *repo.Media) *Fanout {
	return &Fanout{
		Email:   &SMTPSender{Host: cfg.SMTPHost, Port: cfg.SMTPPort, From: cfg.SMTPFrom, User: cfg.SMTPUser, Pass: cfg.SMTPPass},
		Wecom:   &HTTPSender{Name: "WECOM_BOT", URL: cfg.WecomWebhookURL, Mode: cfg.NotifierMode, Media: media},
		Webhook: &HTTPSender{Name: "WEBHOOK", URL: cfg.GenericWebhookURL, Mode: cfg.NotifierMode, Media: media},
	}
}

func (f *Fanout) Send(ctx context.Context, msg Message) error {
	switch msg.Channel {
	case domain.ChannelEmail:
		return f.Email.Send(ctx, msg)
	case domain.ChannelWecom:
		return f.Wecom.Send(ctx, msg)
	case domain.ChannelHook:
		return f.Webhook.Send(ctx, msg)
	default:
		return fmt.Errorf("%w: unknown channel", domain.ErrPermanent)
	}
}

type SMTPSender struct {
	Host, From, User, Pass string
	Port                   int
}

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	to := msg.ToEmail
	if to == "" {
		to = "owner@gopuppy.test"
	}
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	body := strings.Join([]string{
		"From: " + s.From,
		"To: " + to,
		"Subject: " + msg.Title,
		"Content-Type: text/plain; charset=UTF-8",
		"",
		msg.Body,
	}, "\r\n")
	d := net.Dialer{Timeout: 8 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	c, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}
	defer c.Close()
	if s.User != "" {
		if err := c.Auth(smtp.PlainAuth("", s.User, s.Pass, s.Host)); err != nil {
			return fmt.Errorf("%w: %v", reminder.ErrAuthFailure, err)
		}
	}
	if err := c.Mail(s.From); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, body); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

type HTTPSender struct {
	Name  string
	URL   string
	Mode  string
	Media *repo.Media
}

func (h *HTTPSender) Send(ctx context.Context, msg Message) error {
	payload, _ := json.Marshal(map[string]any{
		"channel": h.Name,
		"title":   msg.Title,
		"body":    msg.Body,
		"at":      clock.FormatDateTime(clock.Now()),
	})
	if h.URL == "" || h.Mode == "mock" {
		if h.Media != nil {
			return h.Media.InsertMockDelivery(ctx, h.Name, string(payload), clock.Now())
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	switch resp.StatusCode {
	case 401, 403:
		return fmt.Errorf("%w: %d %s", reminder.ErrAuthFailure, resp.StatusCode, string(b))
	case 422:
		return fmt.Errorf("%w: %d %s", reminder.ErrValidationRemote, resp.StatusCode, string(b))
	}
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return fmt.Errorf("%w: %d %s", domain.ErrTransient, resp.StatusCode, string(b))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: %d %s", domain.ErrPermanent, resp.StatusCode, string(b))
	}
	return nil
}
