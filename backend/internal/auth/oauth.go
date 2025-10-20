package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rishabh-sonic/orbit/pkg/config"
)

type OAuthService struct {
	cfg     *config.Config
	authSvc *Service
}

func NewOAuthService(cfg *config.Config, authSvc *Service) *OAuthService {
	return &OAuthService{cfg: cfg, authSvc: authSvc}
}

// GoogleAuthURL returns the Google OAuth authorization URL.
func (o *OAuthService) GoogleAuthURL(state string) string {
	params := url.Values{
		"client_id":     {o.cfg.GoogleClientID},
		"redirect_uri":  {o.cfg.WebsiteURL + "/auth/google"},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// GitHubAuthURL returns the GitHub OAuth authorization URL.
func (o *OAuthService) GitHubAuthURL(state string) string {
	params := url.Values{
		"client_id":    {o.cfg.GitHubClientID},
		"redirect_uri": {o.cfg.WebsiteURL + "/auth/github"},
		"scope":        {"user:email"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// --- Google ---

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (o *OAuthService) HandleGoogle(ctx context.Context, code string) (string, error) {
	token, err := o.exchangeCode(
		"https://oauth2.googleapis.com/token",
		o.cfg.GoogleClientID, o.cfg.GoogleClientSecret, code,
		o.cfg.WebsiteURL+"/auth/google",
	)
	if err != nil {
		return "", fmt.Errorf("google token exchange: %w", err)
	}
	info, err := fetchJSON[googleUserInfo](
		"https://www.googleapis.com/oauth2/v3/userinfo",
		"Bearer "+token,
	)
	if err != nil {
		return "", fmt.Errorf("google userinfo: %w", err)
	}
	username := sanitizeUsername(info.Name)
	return o.authSvc.GetOrCreateOAuthUser(ctx, "google", info.Sub, info.Email, username, info.Picture)
}

// --- GitHub ---

type githubUserInfo struct {
	ID     int    `json:"id"`
	Login  string `json:"login"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar_url"`
}

func (o *OAuthService) HandleGitHub(ctx context.Context, code string) (string, error) {
	token, err := o.exchangeCode(
		"https://github.com/login/oauth/access_token",
		o.cfg.GitHubClientID, o.cfg.GitHubClientSecret, code,
		o.cfg.WebsiteURL+"/auth/github",
	)
	if err != nil {
		return "", err
	}

	info, err := fetchJSON[githubUserInfo]("https://api.github.com/user", "Bearer "+token)
	if err != nil {
		return "", err
	}

	// GitHub may not return email — fetch separately
	emailAddr := info.Email
	if emailAddr == "" {
		emailAddr = fmt.Sprintf("%d+%s@users.noreply.github.com", info.ID, info.Login)
	}
	name := info.Name
	if name == "" {
		name = info.Login
	}
	return o.authSvc.GetOrCreateOAuthUser(ctx, "github", fmt.Sprintf("%d", info.ID), emailAddr, sanitizeUsername(name), info.Avatar)
}

// --- helpers ---

func (o *OAuthService) exchangeCode(tokenURL, clientID, clientSecret, code, redirectURI string) (string, error) {
	params := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	}
	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
	}

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("oauth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

func fetchJSON[T any](urlStr, authorization string) (T, error) {
	var zero T
	req, _ := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return zero, fmt.Errorf("HTTP %d from %s", resp.StatusCode, urlStr)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, err
	}
	return result, nil
}

func sanitizeUsername(name string) string {
	name = strings.TrimSpace(name)
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		} else if r == ' ' {
			sb.WriteRune('_')
		}
	}
	result := sb.String()
	if len(result) < 3 {
		result = result + "user"
	}
	if len(result) > 30 {
		result = result[:30]
	}
	return strings.ToLower(result)
}
