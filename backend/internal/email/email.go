package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rishabh-sonic/orbit/pkg/config"
)

type Sender struct {
	apiKey    string
	fromEmail string
}

func New(cfg *config.Config) *Sender {
	return &Sender{
		apiKey:    cfg.ResendAPIKey,
		fromEmail: cfg.ResendFromEmail,
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *Sender) Send(to, subject, html string) error {
	if s.apiKey == "" {
		// Log-only in dev when no key is configured
		fmt.Printf("[EMAIL] to=%s subject=%s\n", to, subject)
		return nil
	}

	payload := sendRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}

func (s *Sender) SendVerificationCode(to, code string) error {
	html := fmt.Sprintf(`
<p>Your verification code is:</p>
<h2>%s</h2>
<p>This code expires in 15 minutes.</p>
`, code)
	return s.Send(to, "Verify your email", html)
}

func (s *Sender) SendPasswordReset(to, code string) error {
	html := fmt.Sprintf(`
<p>Your password reset code is:</p>
<h2>%s</h2>
<p>This code expires in 15 minutes. If you did not request a reset, ignore this email.</p>
`, code)
	return s.Send(to, "Reset your password", html)
}

func (s *Sender) SendCommentReplyNotification(to, posterName, postTitle, postURL string) error {
	html := fmt.Sprintf(`
<p><strong>%s</strong> replied to your post <a href="%s">%s</a>.</p>
`, posterName, postURL, postTitle)
	return s.Send(to, "New reply on "+postTitle, html)
}
